## grpc-server example with proto reflection

#### [Optional] Build protobufs for example-grpc-server 

Do it only in case if you want to update client/server stubs or descriptor binary file for grpc. Protofiles were built using buf ([buf installation guide](https://buf.build/docs/installation).

 ```
 cd cmd/examples/grpcserver
 buf generate
 buf build -o gen/user/v1/user.desc
 ```

#### Run example-grpc-server

1. Run example-grpc-server 
   - `make build-example-grpc-server`
   - `./example-grpc-server`

2. Run grpc-gateway
    -  `make build service`
    - `./grpc-rest-proxy`

3. Send http request to grpc-gateway (example)

```
 curl -X GET http://localhost:8080/api/user/John 
```    

Following table shows examples of tested routes patterns on grpc-gateway.

### Route patterns:
  
| Method | Pattern                                                   | Path                                                | additional_bindings                     | query params                          |       body |   |   |   |   |
|:------:|:---------------------------------------------------------:|:---------------------------------------------------:|:---------------------------------------:|:-------------------------------------:|:----------:|:---:|:---:|:---:|:---:|
| GET    | /api/user/{username}                                      | /api/user/john                                      |                                         |                                       | *          |   |   |   |   |
| GET    | /api/users/{username}                                     | /api/users/john                                     | /api/users/{username}/country/{country} |                                       |            |   |   |   |   |
| GET    | /api/users/job/{job.job_title=/*/}                        | /api/users/job/architect                            |                                         |                                       |            |   |   |   |   |
| GET    | /api/users/filter                                         | /api/users/filter?username=john&country=usa         |                                         | ?username=&country=&company=&jobtype= |            |   |   |   |   |
| GET    | /api/users/address/{address.country=/*/}/posts/{type=**}" | /api/users/address/alabama/posts/promotion          |                                         |                                       |            |   |   |   |   |
| GET    | /api/users/summary/{summary=**}                           | /api/users/summary/username/country/company/jobtype |                                         |                                       |            |   |   |   |   |
| POST   | /api/users/create                                         | /api/users/create                                   |                                         |                                       | user       |   |   |   |   |
| DELETE | /api/user/delete/{username=/*/}                           | /api/user/delete/john                               |                                         |                                       |            |   |   |   |   |
| PUT    | /api/user/username/{username}/{job.job_title=*}/*         | /api/user/username/john/architect/country/usa       |                                         |                                       |            |   |   |   |   |
