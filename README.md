```
go run main.go -start 1 -end 100 -url mongodb://localhost:27017
```

## Accessing MongoDB on Remote Server from Local Workstation
```
ssh <host> -N -f -L 27017:localhost:27017

# Terminate port forwarding
ps aux | grep ssh
kill <pid>
```