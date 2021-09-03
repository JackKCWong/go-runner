# gorun

The go-runner-cli for CRUD go apps

## get started

```bash
# install gorun cli
go install github.com/JackKCWong/go-runner/cmd/...
# for go1.14 or below
go get -u github.com/JackKCWong/go-runner/cmd/...

# initialize your app
mkdir your-app
cd your-app
gorun new your-module-name # create an app from example 
git commit -a -m "init commit"
gorun pub # for deploying the 1st time
```
