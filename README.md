
# [Hiding Go HTTP Client Behind a Proxy or Tor](https://medium.com/@tufin/how-to-use-a-proxy-with-go-http-client-cfc485e9f342)

By Effi Bar-She’an and Reuven Harrison

Say you’re a DevOps or a security manager and you want to make sure 
some or maybe all of your pods use *Tor* or some other proxy as an egress gateway.

There are three ways to instruct a client to use a proxy:

1. Set the HTTP_PROXY environment variable:

    	$ export HTTP_PROXY="http://ProxyIP:ProxyPort"

HTTP_PROXY environment variable will be used as the proxy URL for HTTP requests and HTTPS requests, unless overridden by HTTPS_PROXY or NO_PROXY

2. Explicitly instructing the HTTP client to use a proxy. Here’s an example in golang:

	    proxy, _ := url.Parse("http://ProxyIP:ProxyPort") httpClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}

For a more robust *HTTP Client* checkout [this](https://github.com/Tufin/blog/blob/master/go-proxy/common/http.go).

3. Golang also offers the default transport option as part of the “net/http” package. This setting instructs the entire program (including the default HTTP client) to use the proxy:

    	proxy, _ := url.Parse("http://ProxyIP:ProxyPort") http.DefaultTransport := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}

## The Tor Proxy

[Tor](https://www.torproject.org/) aims to defend against tracking and surveillance. If you want to write an application that can’t be traced, an easy solution can be using *Tor* as a proxy.

## Tor Installation

You can use Tor in a few ways:

1. Install [Tor browser](https://tb-manual.torproject.org/installation/) or use another browser like [brave](https://brave.com/) that comes with *Tor* browsing as an option

1. Install as a proxy service on your computer, see [Tor docs](https://2019.www.torproject.org/docs/tor-doc-osx.html.en)

1. Run *Tor* inside a Docker container

## Running Tor inside a Docker container

Running *Tor* inside a Docker container makes it easy if you want to package your application with *Tor*. For example, if you want to run a batch of HTTP calls as part of CI workflow.

### How to?

1. Copy the following to a file: [Dockerfile.tor](https://github.com/Tufin/blog/blob/master/go-proxy/Dockerfile.tor)

	    FROM alpine:edge
	    RUN apk update && apk add tor
	    EXPOSE 9150
	    USER tor
	    CMD ["/usr/bin/tor"]

2. Create a docker image named tor (optional):

    	docker build -t tor -f Dockerfile.tor .

3. Run the docker image you just created:

    	docker run -d --rm --name tor -p 9150:9150 tor

Or use the image from github (if you want to skip 2)

    docker run -d --rm --name tor -p 9150:9150 tufin/tor

After that, you’ll have a *Tor* proxy running on 127.0.0.1:9150 so go ahead and configure your browser to use a SOCKS proxy on 127.0.0.1:9150, or use *Tor* as a proxy for *Go* client.

## Using Tor as a Proxy for Go Client

Like we did above, just replace the URL to the running *Tor*:

    proxy, _ := url.Parse("socks5://127.0.0.1:9050") http.DefaultTransport := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}

For a more robust *HTTP Client* checkout [this](https://github.com/Tufin/blog/blob/master/go-proxy/common/http.go).

## Using Tor as an egress proxy inside a Kubernetes cluster

If your application is running inside a k8s cluster, it would be nice to have an HTTP Tor Proxy, so any internal service can use it. In order to do that let’s combine all the above, and a little more :)

Our architecture will look like this:

![](https://cdn-images-1.medium.com/max/3264/1*OXKvCCIywBqZVuyEP9nSKw.jpeg)

## Deploy a Tor Egress Proxy

The following YAML contains k8s service, and a deployment for *Tor* (same docker image as above):

```yaml
    apiVersion: v1
    kind: Service
    metadata:
      name: tor
      namespace: default
      labels:
        app: tor
    spec:
      selector:
        app: tor
      ports:
        - name: http
          port: 9050
          targetPort: 9050
          protocol: TCP
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: tor
      namespace: default
    ---
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: tor
    spec:
      replicas: 1
      strategy:
        rollingUpdate:
          maxSurge: 50%
          maxUnavailable: 50%
        type: RollingUpdate
      template:
        metadata:
          labels:
            app: tor
        spec:
          serviceAccountName: tor
          containers:
            - name: tor
              image: tufin/tor
              imagePullPolicy: Always
              ports:
                - containerPort: 9050
```

Let’s configure a *Go* service to use our *Tor* Egress proxy service by adding an HTTP_PROXY header so you don't need to use a special Go HTTP client; *Go* client use it by default.

```yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: demo
      namespace: default
    ---
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: demo
    spec:
      replicas: 1
      strategy:
        rollingUpdate:
          maxSurge: 50%
          maxUnavailable: 50%
        type: RollingUpdate
      template:
        metadata:
          labels:
            app: demo
        spec:
          serviceAccountName: demo
          containers:
            - name: demo
              image: myapp
              imagePullPolicy: Always
              env:
                - name: HTTP_PROXY
                  value: socks5://tor:9050
```

## Tor in Action

Now that you configured your proxy, you can see that the *time* service is connecting to its public endpoints through the tor network:

![View from [SecureCloud](https://www.tufin.com/tufin-orchestration-suite/securecloud) Service Graph](https://cdn-images-1.medium.com/max/7164/1*za7ynTq2LK-HN00VrlhO4g.png)*View from [SecureCloud](https://www.tufin.com/tufin-orchestration-suite/securecloud) Service Graph*

You can go one step further and enforce a Kubernetes network policy to restrict specific pods to connect only to the egress proxy and not to any external IP, this will ensure that no pods are bypassing your proxy.

[SecureCloud](https://www.tufin.com/tufin-orchestration-suite/securecloud) can also be used to automatically generate the policy for you by observing the cluster in run-time or during tests (read more on [this article](https://medium.com/@tufin/generating-kubernetes-network-policies-automatically-678ca0411)).

Effi & Reuven.
