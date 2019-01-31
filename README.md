# Practice k8s Controller Custom Resource
## Introduction
To understand how to develop the k8s custom, I follow the code from this repo:
* https://github.com/trstringer/k8s-controller-custom-resource

I also refer to two major articles listed as the followings.
First article explains more detail about CRD, and the second one has example to hand on.
* https://medium.com/@trstringer/create-kubernetes-controllers-for-core-and-custom-resources-62fc35ad64a3

In this example, I also build a gin-gonic base http service, you can refer to
* https://github.com/k2star0118/practice-gin-gonic
And I push the docker image to my dockerhub
* k2star0118/practice-gin-gonic

## Run
```console
// Apply config
$ kubectl apply -f ./crd/myresource.yaml

// Run the CRD
$ go run main.go

// Create a custom resource of type MyResource
// You can see the enable/disable get, put value in this example file
$ kubectl apply -f ./example/example-myresource.yaml
myresource.trstringer.com/example-gin-gonic-http created
```

### Verify
You should get pod ip first, here example is 172.17.0.5
```console
// Get the pod ip
$ kubectl get pods -o wide
NAME                                        READY   STATUS    RESTARTS   AGE     IP           NODE       NOMINATED NODE
example-gin-gonic-http-6bfcb797b-f2w54      1/1     Running   0          15d     172.17.0.5   minikube   <none>
```

To use curl command, you should login your k8s cluster node. If you run minikube, you can login via
```console
$ minikube ssh
```

For this practice, we default enable get only for gin-gonic http service.
```console
$ curl -X GET http://172.17.0.5:8888/example
{"message":"Successfully to query get example"}

$ curl -X PUT http://172.17.0.5:8888/example
"Does not enable put method"
```

When you change the the file "example/example-myresource.ymal" and apply config again,
you will get the different curl result
```console
$ kubectl apply -f ./crd/myresource.yaml
myresource.trstringer.com/example-gin-gonic-http configured
```

If you change the value {spec.someValue} from 1 to 2, then you will get
```console
$ curl -X GET http://172.17.0.5:8888/example
"Does not enable get method"

$ curl -X PUT http://172.17.0.5:8888/example
{"message":"Successfully to query put example"}
```

## Develop step
When you want to deploy your own docker container and do some management via k8s,
you could hand on via following steps.

### 1. Define your own resource manager via yaml file
k8s calls it "Custom Resource Definition" (CRD). In this section, it defines two items:
* Custom Resource Struct Name
* Custom Resource API

You could find the file under "crd" folder, I also add more comments in this file which name
is myresource.yaml. After defined, run following command to create

```console
$ kubectl apply -f ./crd/myresource.yaml
```
Create a custom resource of type MyResource
```console
$ kubectl apply -f ./example/example-myresource.yaml
```

### 2-1. Define your own resource struct
There are 3 files we need to define:
* Package level: /pkg/apis/myresource/v1/doc.go
* Your resource structure: /pkg/apis/myresource/v1/types.go
* API schema register for k8s: /pkg/apis/myresource/v1/register.go

In these files, you not only need to define the data,
but also need to add some comments for the code generator. 
For example, in the resource structure file, you will see some comments like followings. 
```
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
```
These are “indicators” for the code generator, and their meanings are:
* +genclient — generate a client (see below) for this package
* +genclient:noStatus — when generating the client, there is no status stored for the package
* +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object — generate deepcopy logic (required)
implementing the runtime.Object interface (this is for both MyResource and MyResourceList)

### 2-2. Generate custom resource code
Generate the code to interact with k8s for your own resource. It locates in "/pkg/client" folder.
```console
$ sh k8s_ctrl_code_generator.sh
```
After running the code generator we now have generated code that handles a large array of functionality for our new resource.
Now we need to tie a lot of loose ends together for our new resource.

#### Note
When you run the command in step 4, you may face this issue, if you use go module instead of godep
https://github.com/kubernetes/kubernetes/issues/67566

### 3. Write controller to manage CRD
This is the part to define how to manager your resource
* main.go — this is the entry point for the controller as well as where everything is wired up. 
* controller.go — the Controller struct and methods, 
and where all of the work is done as far as the controller loop is concerned
* handler.go — the sample handler that the controller uses to take action on triggered events