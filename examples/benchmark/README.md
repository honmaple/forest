Fork from [github.com/vishr/web-framework-benchmark](github.com/vishr/web-framework-benchmark)

# Web Framework Benchmark

- `go get github.com/honmaple/forest/examples/benchmark`

## Running Benchmark
```
└──╼ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/honmaple/forest/examples/benchmark
cpu: Intel(R) Core(TM) i5-8279U CPU @ 2.40GHz
BenchmarkEchoStatic-8        	   35714	     31852 ns/op	    2225 B/op	     157 allocs/op
BenchmarkEchoGitHubAPI-8     	   23641	     48007 ns/op	    3089 B/op	     203 allocs/op
BenchmarkEchoGplusAPI-8      	  475256	      2508 ns/op	     176 B/op	      13 allocs/op
BenchmarkEchoParseAPI-8      	  261531	      4322 ns/op	     340 B/op	      26 allocs/op
BenchmarkGinStatic-8         	   35372	     32770 ns/op	    8516 B/op	     157 allocs/op
BenchmarkGinGitHubAPI-8      	   23758	     47356 ns/op	   11203 B/op	     203 allocs/op
BenchmarkGinGplusAPI-8       	  437644	      2801 ns/op	     703 B/op	      13 allocs/op
BenchmarkGinParseAPI-8       	  237919	      4885 ns/op	    1393 B/op	      26 allocs/op
BenchmarkForestStatic-8      	   41277	     28831 ns/op	    2095 B/op	     157 allocs/op
BenchmarkForestGitHubAPI-8   	   25910	     46405 ns/op	    2960 B/op	     203 allocs/op
BenchmarkForestGplusAPI-8    	  477165	      2543 ns/op	     176 B/op	      13 allocs/op
BenchmarkForestParseAPI-8    	  251432	      4546 ns/op	     345 B/op	      26 allocs/op
PASS
ok  	github.com/honmaple/forest/examples/benchmark	17.770s
```
