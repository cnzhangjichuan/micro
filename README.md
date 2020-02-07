# Micro
golang for simple micro sevice. supported http, websocket, remote call by socket.

# Start simple service
## Register api
```
// only "admin" permits can calls.
var permits = "admin"

micro.Register("/apitest", permits, func(dpo types.Dpo) error {
    // read data from remote
    var param strut {
        P1 string
        P2 string
    }
    dpo.Request(&r)

    var resp struct {
        PageIndex int
        PageCount int
        List []interface{}
    }
    ....
    // set reponse data
    dpo.Reponse(&resp)

    return nil
})
```

## Start service
```
micro.Service(types.EnvConfig{
    Id:   "service", // service id
    Port: "8088",
})
```

# Start distributed services
## Step1: Start a central server for inter service registration
```
micro.Service(types.EnvConfig{
    Id:   "center",
    Port: "8000",
})
```

## Step2: Start services
```
// service A
micro.Service(types.EnvConfig{
    Id:   "service-a",
    Port: "8001",
    Register: "127.0.0.1:8000"
})

// service B1
micro.Service(types.EnvConfig{
    Id:       "service-b",
    Port:     "8002",
    Register: "127.0.0.1:8000",
})

// service B2
micro.Service(types.EnvConfig{
    Id:       "service-b",
    Port:     "8003",
    Register: "127.0.0.1:8000",
})

...
```

## Calls: A service calls B service API
```
micro.Register("/testname", "", func(dpo types.Dpo) error {
    ...
    var param strut {
        P1 string
        P2 string
    }

    var resp struct {
        PageIndex int
        PageCount int
        List []interface{}
    }
    err := micro.Load(&resp, "service-b", "/apitest", param)
    ...

    return nil
})
```
# Appointment
**1. assets:** static file, can be directly browsed.