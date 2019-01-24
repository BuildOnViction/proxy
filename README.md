# go-crawler-worker-queues
A Simple Crawler Worker Queues in Golang

### Compile
```
make
```

### Run
```
$ ./crawler -n 1 -url http://thanhson1085.github.io -req 1
```

### Test
Test with 1 worker and implement 100 requests
```
$ time ./crawler -n 1 -url http://thanhson1085.github.io -req 100
real    0m3.206s
user    0m0.009s
sys 0m0.088s

```
Test with 10 worker and implement 100 requests
```
$ time ./crawler -n 10 -url http://thanhson1085.github.io -req 100
real    0m0.488s
user    0m0.064s
sys 0m0.133s
```

The result is amazing!!!

### References
[Worker-Queues-In-Go](http://nesv.github.io/golang/2014/02/25/worker-queues-in-go.html)
