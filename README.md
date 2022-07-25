# Blah Server


## Managing Docker

First things first lets build the dockermgr executable

Download deps
```bash
go mod download && go mod verify
```

Build Executable
```bash
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
```
**Command:** **container**

<small>-image flag and -name flag is required (unless -rm is used)</small>
```
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
```bash
./dockermgr image -context ./ -Name test:1.0.0
```

*Create container and run it*
```bash
./dockermgr container -image test:1.0.0 -Name test_blah1
```

*Print logs to stdout and follow*
```bash
./dockermgr logs -name test_blah1
```

*Remove Image (Forcefully)*

```bash
./dockermgr container -rm test_blah1

```