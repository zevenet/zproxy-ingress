
<a name="summary"></a>  
# zproxy-ingress

[![Go report card](https://goreportcard.com/badge/github.com/zevenet/zproxy-ingress)](https://goreportcard.com/report/github.com/zevenet/zproxy-ingress)
![License](https://img.shields.io/github/license/zevenet/kube-nftlb)

`zproxy-ingress` is an ingress controller for Kubernetes. It connects with kubernetes API and configures a
[`zproxy`](https://github.com/zevenet/zproxy) daemon to route the HTTP requests that income in the Kubernetes cluster.

This controller needs some permissions to read information from the Kubernetes API and create the zproxy configuration. Such information is about:
* ingresses rules: to create the zproxy services.

* configmaps: to configure the global and defaults zproxy settings.

* secrets: to create the SSL certificates for HTTPS incoming requests.

The zproxy-ingress container contains two daemons, a GO client that connects with the k8s API and a zproxy daemon that manages the incoming requests and forwards them to the k8s services.

## Index

1. [ Starting ](#starting)  
	1.1 [Build](#build)</br>
	1.2 [Some commands to inspect the pod and container](#build)
	
2. [ Settings ](#settings)  
	2.1 [Container configuration files](#configfiles)</br>
	2.2 [Environment varibles](#environment)</br>
	2.3 [ConfigMaps](#configmaps)</br>
	2.4 [Annotations](#annotations)
	
3. [Some notes ](#someNotes)  
	3.1 [Ingress Namespace](#ingressNamespace)</br>
	3.2 [How to link an ingress rule with zproxy-ingress](#linkAnIngressRule)</br>
	3.3 [How to set up the default ingress controller](#defaultIngressController)</br>
	3.4 [How SSL certificates work](#certificatesWork)</br>
	3.5 [How to configure SSL certificates](#configureCertificates)</br>
	3.6 [How to configure the default SSL certificate](#configureDefaultCertificate)</br>
	3.7 [Configuring DH param](#configureDHparam)</br>
	3.8 [Load balancing among ingress backends](#loadBalancingAmongIngress)</br>
	3.9 [Define a redirect in an ingress rule](#defineRediret)</br>
	3.10 [Default backend](#defaultBackend)

4. [ Contributing ](#contributing)  
5. [ Authors ](#authors)  

<a name="starting"></a>  
## Starting :rocket:

<a name="build"></a>  
### Build

This project depends on ZEVENET CE repository to create the container with the latest zproxy package. So, this project
contains only the code for the Kubernetes GO client and to create the docker container, it does not manage any zproxy source code or compilation.

```shell
root@k8s:~# git clone https://github.com/zevenet/zproxy-ingress
root@k8s:~# cd zproxy-ingress
root@k8s:~# ./build.sh
root@k8s:~# kubectl apply -f yaml/
```

<a name="commands"></a>  
### Some commands to inspect the pod and container

* Get the logs of the pod:

```shell
root@k8s:~# kubectl logs `kubectl get pods -n zproxy-ingress | grep zproxy | grep -Ev Terminating | cut -d " " -f1` -n zproxy-ingress
```

* Get the zproxy configuration file:

```shell
root@k8s:~# kubectl exec -it `kubectl get pods -n zproxy-ingress | grep zproxy-ingress | grep -Ev Terminating | cut -d " " -f1` -n zproxy-ingress -- cat  /run/ingress.cfg
```

* Check if all objects where created properly:

```shell
root@k8s:~# kubectl logs `kubectl get pods -n zproxy-ingress | grep zproxy | grep -Ev Terminating | cut -d " " -f1` -n zproxy-ingress | grep "Latest reload was: " | tail -1 | sed 's/Latest reload was: //'
```


<a name="settings"></a>  
## Settings :gear:

<a name="configfiles"></a>  
### Container configuration files

Those configuration files defined the configuration that the container is configured. This configuration can be modified through the different k8s features (ENV, configmaps and annotations).

*  Paths: paths to deploy configuration files and binaries in the container.

*  Client: parameters to the GO client that manages the zproxy daemon. They can be modified using ENV variables.

*  Global: global parameters of the zproxy daemon, they only can be modified in the container start using ENV variables.

*  Listener: parameters for the HTTP and HTTPS listeners. They can be modified using configmap.

*  Service: parameters for a zproxy services. The default values can be modified using configmap or can be specified for a service through annotations.

<a name="environment"></a>  
### Environment variables

Some examples can be checked in the test: tests/010_env_params/

Those variables cannot be modified in the container runtime, it only is configured in the container starting time. They are variables for the GO client and the zproxy daemon.

Variables for the GO client
* DaemonsCheckTimeout: This time is the interval (in seconds) to check that both processes are alive, else the container will be closed.

* ClientLogsLevel: It is the debug log level for the zproxy-ingress GO client. More log level will include more log details. (Values: 0, 1 or 2 ).

* ClientStartGraceTme: It is the time in seconds to wait between two client controllers starting. It is useful to avoid race conditions when a controller depends on another. For example, before applying services rules it is needed to load the certificates.

Variables for the global configuration of the zproxy daemon.

* ProxyLogsLevel: It is the zproxy daemon log level.

* DHFile: Use the supplied dhparams pem file for DH key exchange for non-export-controlled negotiations.  Generate such a file with openssl dhparam.

* ECDHCurve: Use for the listeners the named curve for elliptical curve encryption.

* TotalTO: How long should zproxy wait for a response from the back-end (in seconds).

* ConnTO: How long should zproxy wait for a connection to the back-end (in seconds).

* AliveTO: Specify how often zproxy will check for resurrected back-end hosts.

* ClientTO: Specify how long zproxy will wait for a client request. After this long has passed without the client sending any data zproxy will close the connection.

* Ignore100Continue: If it is set with value **1**, zproxy will ignore the "Expect: 100-continue" header. If it is set with **0** (by default) zproxy will manage "Expect: 100-continue" headers.

<a name="configmaps"></a>  
### ConfigMaps	

Some examples can be checked in the test: tests/009_configmap/

The listener parameters will be used for all the incoming requests.

* listener-http-port: It is the port used by the controller to listen to the HTTP request. The port used by default is **80**.

* listener-https-port: It is the port used by the controller to listen to the HTTPS request. The port used by default is **443**.

* listener-error-414: It is a string with *HTML* syntaxis used as the reponse in the case that proxy will return the HTTP error code 414.

* listener-error-500: It is a string with *HTML* syntaxis used as the reponse in the case that proxy will return the HTTP error code 500.

* listener-error-501: It is a string with *HTML* syntaxis used as the reponse in the case that proxy will return the HTTP error code 501.

* listener-error-503: It is a string with *HTML* syntaxis used as the reponse in the case that proxy will return the HTTP error code 503.

* listener-xhttp: It is the set of HTTP verbs that the proxy will accept. The available options are:
    * 0 (default) accepts only standard HTTP requests (GET, POST, HEAD).
    * 1 accepts the above methods and additionally allow extended HTTP requests (PUT, PATCH, DELETE).
    * 2 accepts the above methods and additionally allow standard WebDAV verbs (LOCK, UNLOCK, PROPFIND, PROPPATCH, SEARCH, MKCOL, MOVE, COPY, OPTIONS, TRACE, MKACTIVITY, CHECKOUT, MERGE, REPORT).
    * 3 accepts the above methods and additionally allow MS extensions WebDAV verbs (SUBSCRIBE, UNSUBSCRIBE, NOTIFY, BPROPFIND, BPROPPATCH, POLL, BMOVE, BCOPY, BDELETE, CONNECT).
    * 4 accepts the above methods and additionally allow MS RPC extensions verbs (RPC_IN_DATA, RPC_OUT_DATA).

* listener-rewrite-location:  If  1 force zproxy to change the Location: and Content-location: headers in responses. If they point to the back-end itself or the listener (but with the wrong protocol) the response will be changed to show the virtual host in the request. Default:
              1 (active).  If the value is set to 2 only the back-end address is compared; this is useful for redirecting a request to an HTTPS listener on the same server as the HTTP listener.

* listener-remove-request-header: It removes  certain  headers  from  the incoming requests (request from the client). All occurrences of the matching specified header will be removed.
              Multiple directives may be specified (using new line separator "\n") in order to remove more than one header, and the header itself may be a regular pattern (though this should be used with caution).

* listener-remove-response-header: It removes  certain  headers  from  the outcoming requests (response from the backend). All occurrences of the matching specified header will be removed.
              Multiple directives may be specified (using new line separator "\n") in order to remove more than one header, and the header itself may be a regular pattern (though this should be used with caution).

* listener-add-request-header: It adds the defined header to the request passed to the back-end server. The header is added verbatim. Several headers can be added using the new line separator "\n".

* listener-add-response-header: It adds the defined header to the response passed to the client. The header is added verbatim. Several headers can be added using the new line separator "\n".
* listener-default-cert: It is the certificate used for HTTPS requests when the incoming request does not match with any TLS certificate.
* listener-ciphers: This is the list of ciphers that will be accepted by the SSL connection; it is a string in the same format as in OpenSSL ciphers(1) and SSL_CTX_set_cipher_list(3).
* listener-disable-ssl-protocol: Disable the protocol and all lower protocols as well. This is due to a limitation in OpenSSL, which does not support disabling a single protocol. For example, Disable TLSv1 would disable SSLv2, SSLv3 and TLSv1, thus allowing only TLSv1_1 and TLSv1_2.
              [NOTE]Disable TLSv1_3 would disable only TLSv1_3.
              The options are: SSLv2, SSLv3, TLSv1, TLSv1_1, TLSv1_2 or TLSv1_3
* listener-ssl-honor-cipher-order: If this field is 1, the server will broadcast a preference to use Ciphers in the order supplied in the Ciphers directive. If the value is 0, the server will treat the Ciphers list as the list of Ciphers it will accept, but no preference will be indicated. The default value is 1.


The service parameters will be used as the default configuration for services. They can be overwritten for each ingress rule using the annotations:

* service-https-backends: It uses SSL to encrypt the request when it is sent to the backend.

* service-strict-transport-security: It is the time in seconds for the StrictTransportSecurity header. By default *21600000*.

* service-cookie-name: Cookie parameters is used to insert a cookie in the HTTP communications with the client. The cookie name (session ID) will be used for identifying the sticky process to backends. This parameter has to be defined as different from an empty string "" to insert the cookie.

* service-cookie-path: It manages the cookie path value for the given cookie.

* service-cookie-domain: Cookie insertion will be executed if the domain matches in the cookie content.

* service-cookie-ttl: It is the time to live in seconds of the cookie, the cookie will be expired after this time.

* service-redirect-url: This parameter it a URI to redirect the client as the response for a service, when it is used, the backend parameter of an ingress rule is ignored. To avoid this behavior for each service, it cannot be set using configmap, it is set up through annotation in the rule which will be required.

* service-redirect-code: It is the HTTP code returned with the redirect, the options are: 301, 302 (by default) or 307. It is not used while RedirectURL is not configured.

* service-redirect-type: It is how the redirect URL is done. The options are: **default** to send the same URL defined in the RedirectURL field, or **append** to append the incoming URL to the -   service-redirect-url field. It is not used while -   service-redirect-url is not configured.

	Examples: if you specified `Redirect "http://abc.example"` and the client requested `http://xyz/a/b/c` then it will be redirected to `http://abc.example`, but if you specified `RedirectAppend "http://abc.example"` it will be sent to `http://abc.example/a/b/c`.

* service-session-type: This parameter defines how the HTTP service is going to manage the client session. The options are:
	*  **""** empty string, no action is taken and persistence is not activated;
	* **IP** the persistence session is done based on the client IP;
	*  **BASIC** the persistence session is done based on the BASIC headers;
	* **URL** the persistence session is done based on a field in the URI;
	*  **PARM** the persistence session is done based on a value separated by “;” at the end of the URI;
	* **COOKIE** the persistence session is done based on a cookie name, this cookie has to be created by the backends; and
	* **HEADER**, the persistence session is done based on a Header name.

* service-session-id: It is available if the persistence field is URL, COOKIE or HEADER, the parameter value will be searched by the profile in the HTTP header and will manage the client session.

* service-session-ttl: The time to live for an inactive client session (max session age) in seconds.


<a name="annotations"></a>  
### Annotations

They are used to overwritten the global service configuration. The selected domain for zproxy annotations is "**`zproxy.ingress.kubernetes.io/`**".

Some example can be checked in the test: tests/008_annotations/

The description about each annotation can be checked in the [configmap](https://github.com/zevenet/zproxy-ingress/###configmap) section, and the accepted annotations are the following ones:

* `zproxy.ingress.kubernetes.io/service-https-backends`
* `zproxy.ingress.kubernetes.io/service-strict-transport-security`
* `zproxy.ingress.kubernetes.io/service-cookie-name`
* `zproxy.ingress.kubernetes.io/service-cookie-path`
* `zproxy.ingress.kubernetes.io/service-cookie-domain`
* `zproxy.ingress.kubernetes.io/service-cookie-ttl`
* `zproxy.ingress.kubernetes.io/service-redirect-url`
* `zproxy.ingress.kubernetes.io/service-redirect-code`
* `zproxy.ingress.kubernetes.io/service-redirect-type`
* `zproxy.ingress.kubernetes.io/service-session-type`
* `zproxy.ingress.kubernetes.io/service-session-id`
* `zproxy.ingress.kubernetes.io/service-session-ttl`


<a name="someNotes"></a>  
## Some notes  :memo:

<a name="ingressNamespace"></a>  
### Ingress Namespace

An ingress rule has assigned a namespace. The resources (secrets and services) that the rule manages have to be in the same namespace than the ingress rule

<a name="linkAnIngressRule"></a>  
### How to link an ingress rule with zproxy-ingress

Zproxy-ingress will configure the ingress rules that contain the incressClassName **zproxy-ingress**.
It should be defined in the spec context:

```yml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: 004-complex-ruleset
spec:
  ingressClassName: zproxy-ingress
  ...
```

<a name="defaultIngressController"></a>  
### How to set up the default ingress controller

A Kubernetes IngressClass object is required to set the zproxy-ingress as the default one.

The configuration yaml **yaml/default-ingressclass.yaml** can be used for this porpuse.

```yml
apiVersion: networking.k8s.io/v1beta1
kind: IngressClass
metadata:
  name: zproxy-ingress
  annotations:
    ingressclass.kubernetes.io/is-default-class: "true"
spec:
  controller: zevenet/ingress-controller
```

<a name="certificatesWork"></a>  
### How SSL certificates work

HTTPS listener will be configured in the ingress rules that contain the **tls** field.

Zproxy-ingress will ignore the field *hosts* to apply the certificate. The certificate will be applied using its own domain parameter, if the domain defined in the certificate matches with an HTTPS request, the certificate that matches first will be used.

The certificates are added to the listener in the same order than appear in the ingress rules. So the certificates more specific should be applied before than more general. I.E.: A certificate for the domain *.internal.zevenet .com
 should appear above than  *zevenet.com in the tls certificates list.

Before than apply a certificate to a ingress rule, it  should be created, else it doesn't be apply until a ingress object will be added/modified/deleted.

```yml
spec:
  tls:
  - hosts:
    - 005.tls
    secretName: tls-cert
  - hosts:
    - 005.pem
    secretName: pem-cert
  - hosts:
    - 005.not.found
    secretName: tls-cert-2
```


<a name="configureCertificates"></a>  
### How to configure SSL certificates

SSL certificates have been integrated with Kubernetes  using the secret feature. Zproxy uses certificates with *pem* format, for that reason, the secrets should be created from a PEM file or using Kubunernetes *TLS* secrets (then, zproxy-ingress client will create the *pem* certificate).

NOTE:
* The certificate name will be the secret name.
* The key used to save the secret  is important to specify the certificate format.
* The certificate should be in the same namespace as the ingress rule which will be linked.

Creating a secret from a *pem* file (notice that is saved in a secret with **pem** key):
```shell
root@k8s:~# kubectl create secret generic <secretname> --from-file=pem=./zencert.pem [-n namespace]
```
Creating an autosigned certificate and a secret (notice that is saved in a secret adding the **key** and the **cert**  files):
```shell
root@k8s:~# openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout <keyfile> -out <certfile> -subj "/=domain.com/OCN=domain.com"
root@k8s:~# kubectl create secret tls <secretname> --key <keyfile> --cert <certfile> [-n namespace]
```

<a name="configureDefaultCertificate"></a>  
### How to configure the default SSL certificate

To define a certificate as the default one when any other is used for an HTTP request, it has to comply with the following requirements:
* It should be created using the wildcard domain '*'.
* It has to be defined in the "**zproxy-ingress**" namespace.

<a name="configureDHparam"></a>  
### Configuring DH param

The DH params file used to encrypt the HTTPS communication has been included in the container for saving time in the pod deployment, but, it is recommended to generate a new one and linked it with the zproxy-ingress configuration.

The following command can be executed to create a DH param file:

```shell
root@k8s:~# openssl dhparam -5 -out /tmp/dh2048.pem 2048
```

 Once it is created, it should be copied in the Kubernetes node and linked it in the zproxy daemonset yaml mounting the file.
 The commented blocks of the file "tests/010_env_params/daemonset.yaml" show an example.

 The ENV variable **DHFile** (daemontset.yaml file) can be modified in order to change the DH param path inside the container.


<a name="loadBalancingAmongIngress"></a>  
### Load balancing among ingress backends

Zproxy-ingress will load balance the incoming requests among different backends if there are several ingress rules with the same match information (host and URI).
Note that the services should be in the same namespace as the ingress rule.

Yaml example:

```yml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress-lb-svcs
spec:
  ingressClassName: zproxy-ingress
spec:
  rules:
  - host: lb.zevenet.com
    http:
      paths:
      - path: /
        backend:
          serviceName: svc-app-v1
          servicePort: 80
      - path: /
        backend:
          serviceName: svc-app-v2
          servicePort: 80
```

<a name="defineRediret"></a>  
### Define a redirect in an ingress rule

A redirect can be configured as the response to a client in an ingress rule. If the redirect is configured, the backend struct from the ingress rule will be ignored.

The following annotations define the behavior of the redirect action:

* zproxy.ingress.kubernetes.io/service-redirect-url: it is the URL to redirect the client.
* zproxy.ingress.kubernetes.io/service-redirect-code: it is the HTTP code to send in the response.
* zproxy.ingress.kubernetes.io/service-redirect-type: it specifies how to create the response URL regarding the incoming one. See the configmap "service-redirect-type" parameter for further information.


<a name="defaultBackend"></a>  
### Default backend

Zproxy-ingress allows adding a default backend that will respond in case that any rule matches the incoming request (the *path* and *host* fields are not achieved).

The default backend is defined in the spect context without any path and host sets.

Only one default backend will be configured in the case in which several ingress rules include this struct.

Here is shown a yaml example.

```yml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: default-bck
spec:
  ingressClassName: zproxy-ingress

spec:
  backend:
    serviceName: default-svc
    servicePort: 80
```

<a name="contributing"></a>  
## Contributing :clap:

**Pull Requests are WELCOME!** Please submit any fixes or improvements:

* [Project Github Home](https://github.com/zevenet/zproxy-ingress)
* [Submit Issues](https://github.com/zevenet/zproxy-ingress/issues)
* [Pull Requests](https://github.com/zevenet/zproxy-ingress/pulls)

<a name="authors"></a>  
## Authors  :nerd_face:

ZEVENET Team

