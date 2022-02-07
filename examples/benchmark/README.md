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
BenchmarkEchoStatic-8        	   34522	     32306 ns/op	    2259 B/op	     157 allocs/op
BenchmarkEchoGitHubAPI-8     	   24789	     48676 ns/op	    3021 B/op	     203 allocs/op
BenchmarkEchoGplusAPI-8      	  429786	      2746 ns/op	     184 B/op	      13 allocs/op
BenchmarkEchoParseAPI-8      	  249092	      4829 ns/op	     347 B/op	      26 allocs/op
BenchmarkGinStatic-8         	   33598	     33378 ns/op	    8568 B/op	     157 allocs/op
BenchmarkGinGitHubAPI-8      	   24686	     47990 ns/op	   11148 B/op	     203 allocs/op
BenchmarkGinGplusAPI-8       	  432154	      2747 ns/op	     704 B/op	      13 allocs/op
BenchmarkGinParseAPI-8       	  220762	      5131 ns/op	    1405 B/op	      26 allocs/op
BenchmarkForestStatic-8      	   31146	     37876 ns/op	    3935 B/op	     109 allocs/op
BenchmarkForestGitHubAPI-8   	   29334	     45882 ns/op	   13428 B/op	      52 allocs/op
BenchmarkForestGplusAPI-8    	  569188	      2224 ns/op	     702 B/op	       4 allocs/op
BenchmarkForestParseAPI-8    	  242726	      5070 ns/op	    1722 B/op	      14 allocs/op
PASS
ok  	github.com/honmaple/forest/examples/benchmark	17.770s
```
