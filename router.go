package routes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	ParameterRune       = ':'
	GreedyParameterRune = '*'
	Key                 = key("parameters")
)

type key string

type Router struct {
	NotFoundHandler http.Handler

	nodes         map[string]*Router
	names         map[string]string
	parameterized bool
	greedy        bool
	handler       http.Handler
	parameters    []string
	name          string
	pattern       string
}

func (router *Router) IsEnd() bool {
	return len(router.nodes) == 0
}

func (router *Router) update(path []string, node *Router, name string) {
	var parameterized bool
	var greedy bool

	if strings.HasPrefix(path[0], string(ParameterRune)) {
		parameterized = true
		router.parameters = append(router.parameters, strings.TrimPrefix(path[0], ":"))
		path[0] = ""
	} else if strings.HasPrefix(path[0], string(GreedyParameterRune)) {
		router.greedy = true
		router.parameters = append(router.parameters, strings.TrimPrefix(path[0], "*"))
		router.handler = node.handler
		return
	}

	if len(path) > 0 {
		var tree *Router
		var ok bool

		if len(path) > 1 {
			if tree, ok = router.nodes[path[0]]; !ok {
				tree = New()
				tree.parameterized = parameterized
				tree.greedy = greedy
				tree.parameters = router.parameters

				router.nodes[path[0]] = tree
				router.names[name] = tree.pattern
			}

			tree.update(path[1:], node, name)
		} else {
			node.parameterized = parameterized
			node.greedy = greedy
			node.parameters = append(router.parameters, node.parameters...)
			router.nodes[path[0]] = node
			router.names[name] = node.pattern

			tree = node
		}

		for name, pattern := range tree.names {
			if tree.name != "" {
				name = fmt.Sprintf("%s:%s", tree.name, name)
			}

			if tree.pattern != "" {
				pattern = strings.TrimPrefix(pattern, "/")
				pattern = fmt.Sprintf("%s/%s", tree.pattern, pattern)
			}

			router.names[name] = pattern
		}
	}
}

func (router *Router) Resolve(path string) (http.Handler, []string) {
	var parameters = make([]string, 0)
	var node = router

	if path[0] == '/' {
		path = path[1:]
	}

	for path != "" {
		var index = strings.IndexRune(path, '/')
		var part string

		if index == -1 {
			part = path
			path = ""
		} else if node.greedy {
			parameters = append(parameters, path)
			return node.handler, parameters
		} else {
			part = path[:index]
			path = path[index+1:]
		}

		if nextNode, ok := node.nodes[part]; ok {
			node = nextNode
			continue
		} else if nextNode, ok := node.nodes[""]; ok && nextNode.parameterized {
			node = nextNode
			parameters = append(parameters, part)
			continue
		} else if node.NotFoundHandler != nil {
			return node.NotFoundHandler, nil
		} else {
			return nil, nil
		}
	}

	if node.handler != nil && node.IsEnd() {
		return node.handler, parameters
	}

	return nil, nil
}

func (router *Router) Add(path string, handler http.Handler, name string) {
	var node = New()
	node.name = name
	node.handler = handler
	node.pattern = fmt.Sprintf("/%s", strings.Trim(path, "/"))

	router.update(strings.Split(strings.Trim(path, "/"), "/"), node, name)
}

func (router *Router) AddRouter(prefix string, node *Router, namespace string) {
	node.name = namespace
	node.pattern = fmt.Sprintf("/%s", strings.Trim(prefix, "/"))

	router.update(strings.Split(strings.Trim(prefix, "/"), "/"), node, namespace)
}

func (router *Router) Reverse(name string, parameters ...string) (string, error) {
	var buffer bytes.Buffer

	if path, ok := router.names[name]; !ok {
		return "", errors.New("Name not found!")
	} else {
		if path[0] == '/' {
			path = path[1:]
		}

		for path != "" {
			buffer.WriteRune('/')

			var index = strings.IndexRune(path, '/')
			var part string

			if index == -1 {
				part = path
				path = ""
			} else {
				part = path[:index]
				path = path[index+1:]
			}

			if part[0] == ParameterRune || part[0] == GreedyParameterRune {
				buffer.WriteString(parameters[0])

				if part[0] == GreedyParameterRune {
					break
				}

				parameters = parameters[1:]
			} else {
				buffer.WriteString(part)
			}
		}
	}

	return buffer.String(), nil
}

func (router *Router) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if handler, parameters := router.Resolve(request.URL.Path); handler != nil {
		var ctx = context.WithValue(request.Context(), Key, parameters)

		handler.ServeHTTP(response, request.WithContext(ctx))
	} else {
		response.WriteHeader(http.StatusNotFound)
	}
}

func New() *Router {
	return &Router{
		nodes: make(map[string]*Router),
		names: make(map[string]string),
	}
}
