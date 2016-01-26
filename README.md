#Example of simple golang data tracking 
Here I made a prototype of tracker on golang. 
Lua script has been used for validation, which realisation is under NDA, but here u can see powerful example which maintance about 6000 concurent requests.
Results are really depend on environment. Here are tests on 2.7Hz processor with 8 cores:

```
vzhiltsov@dev:~$ ab -t 30 -r -n 1000000 -c 20000 'http://localhost:8080/?ad_id=188332&site_id=8572&creative_id=88888&ip=54.159.55.145&ua=Opera%20Tablet%2011%20Android%202.1---Opera/9.80%20(Android%202.1;%20Linux;%20Opera%20Tablet/ADR-1107051709;%20U;%20xx-xx)%20Presto/2.8.149%20Version/11.10'
This is ApacheBench, Version 2.3 <$Revision: 1604373 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/
 
Benchmarking localhost (be patient)
Completed 100000 requests
Finished 174295 requests
 
 
Server Software:        
Server Hostname:        localhost
Server Port:            8080
 
Document Path:          /?ad_id=188332&site_id=8572&creative_id=88888&ip=54.159.55.145&ua=Opera%20Tablet%2011%20Android%202.1---Opera/9.80%20(Android%202.1;%20Linux;%20Opera%20Tablet/ADR-1107051709;%20U;%20xx-xx)%20Presto/2.8.149%20Version/11.10
Document Length:        36 bytes
 
Concurrency Level:      20000
Time taken for tests:   30.472 seconds
Complete requests:      174295
Failed requests:        0
Total transferred:      27712905 bytes
HTML transferred:       6274620 bytes
Requests per second:    5719.77 [#/sec] (mean)
Time per request:       3496.644 [ms] (mean)
Time per request:       0.175 [ms] (mean, across all concurrent requests)
Transfer rate:          888.13 [Kbytes/sec] received
 
Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0   36  95.6      0     394
Processing:   218 3200 1001.8   3184   11544
Waiting:      218 3198 1001.4   3184   11544
Total:        376 3235 983.5   3206   11545
 
Percentage of the requests served within a certain time (ms)
  50%   3206
  66%   3488
  75%   3699
  80%   3878
  90%   4420
  95%   4706
  98%   5569
  99%   5915
 100%  11545 (longest request)
vzhiltsov@dev:~$
```