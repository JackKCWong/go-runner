# go-runner

A naive imitation of [app-runner](https://github.com/danielflower/app-runner) for Go app ([example](https://github.com/JackKCWong/go-runner-hello-world)).


## apis 

* `GET /api/health` return health of go-runner
  
* `POST /api/apps` register and deploy a go-app
 
    `gitUrl` - git url to the app being deployed
 
    `app` - app name
  
* `PUT /api/:app` operate a go-app
 
    `action` - `deploy` or `restart`

* `GET /api/:app/stdout` stream app stdout 

* `GET /api/:app/stderr` stream app stderr

* `ANY /:app/*` access go-apps


## TODO

* [x] basic app CRUD
* [x] unixsocket support in the back
* [x] endpoint for streaming stdout/stderr of an app 
* [ ] a cli client [gorun](https://github.com/JackKCWong/go-runner/tree/main/cmd/client/gorun)
    * [x] init
    * [x] register
    * [x] push
    * [x] delete
    * [ ] status
    * [ ] curl
* [ ] https support in front
* [ ] tcp socket support in the back
* [ ] try using Namespace to isolate apps (ref: [Linux Namespace](https://medium.com/@teddyking/linux-namespaces-850489d3ccf))
    * [ ] PID namespace
    * [ ] filesystem
* [ ] try using cgroup to manage apps resources (ref: [cgroup](https://github.com/containerd/cgroups))
    * [ ] cpu
    * [ ] memory
