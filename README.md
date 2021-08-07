# go-runner

A naive imitation of [app-runner](https://github.com/danielflower/app-runner) for Go app ([example](https://github.com/JackKCWong/go-runner-hello-world)).

## apis 

* `GET /api/health` return health of go-runner
  
* `POST /api/:app` register and deploy a go-app
  
    `gitUrl` - git url to the app being deployed
  
    `app` - app name
  
* `PUT /api/:app` operate a go-app
  
    `app` - app name
  
    `action` - `deploy` or `restart`

* `ANY /:app/*` access go-apps


## TODO

* [x] basic app CRUD
* [x] unixsocket support in the back
* [ ] endpoint for streaming stdout/stderr of an app 
* [ ] https support in front
* [ ] http support in the back
* [ ] try using Namespace to isolate apps (ref: [Linux Namespace](https://medium.com/@teddyking/linux-namespaces-850489d3ccf))
    * [ ] PID namespace
    * [ ] filesystem
    * [ ] cpu
  