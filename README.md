This is about managing and scaling a 3rd party worker application in order to compute a sequence of random numbers within a defined time constraint.   
It uses the enclosed worker binary to compute 150 data samples. This program spawn and manage multiple worker (child) processes.  
The worker provides a http streaming interface to provide the computed numbers with a throughput of 1 rnd/sec.   
There is only one connection per worker. A worker might crash after some time.

### Key Points
* Write a component to manage the worker processes
* handle the http connections and consume the generated numbers
* collect a total amount of 150 data samples
* compute the total count of numbers you processed
* compute the total time spent
* [optional] scale the system to compute 150 numbers within 10sec
* [optional] scale the system with a max of 16 workers


### The Worker
The worker can be find enclosed (`bin/worker`), it's a compiled go application.

### Start
`./worker -workerId 1 -port 3001`

### API
`http://localhost:3001/rnd?n=100`

`n` defines the number of random numbers the worker will generate.

The endpoint has a throughput of 1 rnd/sec. It's a chunked http stream.

Example Response
```bash
HTTP/1.1 200 OK
Cache-Control: no-cache
Connection: keep-alive
Content-Type: application/text
X-Worker-Id: 1
Date: Wed, 28 Aug 2019 21:30:01 GMT
Transfer-Encoding: chunked

rnd=97
rnd=18
rnd=21
rnd=95
rnd=33
```

### How It Works  
Config params are provided in `WorkerHandler` struct. It launches 16 workers, starting from port # 3001 & so on in a separate goroutine.
Since OS may take some time to launch given worker binary(e.g worker.mac), handler waits till `PauseInterval` and then makes an HTTP request to parse data. 
Behind the scenes, it opens TCP socket connection to each worker binary on separate port. In some cases, process might take longer to start (e.g. OS busy), 
so it will retry HTTP request `MaxRetry` times with a configured `PauseInterval`. If TCP connection is successful, it will read each line and parse number. 
Each number is then passed to a data `channel`, where it will append parsed number to `dataSamples` array. It `break`s from `for loop` when it has 
populated 150 data samples, eventually It returns `Result` with `DataSamples` & `TotalTimeSpent`. 

This whole process of parsing 150 numbers along with launching 16 workers happens within 10 seconds. This could be seen on terminal.