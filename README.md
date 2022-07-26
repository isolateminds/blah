# Blah Server


## Managing Blah Server Container

First things first lets build the dockermgr executable or whatever.

Download deps
```
go mod download && go mod verify
```

Build Executable
```
go build -o ./dockermgr ./src/cmd/docker/main.go
```

Now you have a few commands you can use

```
Commands:
    image       Create an image for the server
    container   Manage containers for the server
    logs        Print Logs to stdout.
```

**Command:** **image**


<small>-context flag is required</small>

```
-context string
    Directory where Dockerfile resides
-name string
    A tag to use as an image name for containers to use (default "blah")
-env string
    Name of the .env file prefix to load
```
**Command:** **container**

<small>-image flag and -name flag is required (unless -rm is used)</small>
```
-expose string
    Expose port number and protocol in the format 80/tcp
-hostname string
    Container hostname (default "blah-server")
-image string
    A tag to use as an image name for containers to use
-name string
    A tag to use as an image name for containers to use
-rm string
    Name of the container to remove
```

**Command:** **logs**

<small>-name flag is required</small>

```
-name string
    Name of the container to print logs from
```

### Example Usage

*Create image and tag it*
```
./dockermgr image -context ./ -Name test:1.0.0 -env test
```

*Create container and run it*
```
./dockermgr container -image test:1.0.0 -Name test_blah1
```

*Print logs to stdout and follow*
```
./dockermgr logs -name test_blah1
```

*Remove Image (Forcefully)*

```
./dockermgr container -rm test_blah1

```
### About **-env**...  <i>whats the point?</i>

the command **image** has a sub-command **-env**. This will be the (prefix) of a **.env** file EG. **.(prefix).env**.

The file will be loaded and used for the created image allowing to logically seperate different parts of the application to its own container.