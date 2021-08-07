# go-runner

A simplistic and naive imitation of [app-runner]() for go apps.

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

 * [ ] https support in front
 * [x] unixsocket support in the back
 * [ ] http support in the back
 * [ ] try using Namespace to isolate apps (ref: [Linux Namespace](https://medium.com/@teddyking/linux-namespaces-850489d3ccf))
    * [ ] PID namespace
    * [ ] filesystem
    * [ ] cpu
  