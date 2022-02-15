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
BenchmarkEchoStatic-8        	   35674	     32620 ns/op	    2226 B/op	     157 allocs/op
BenchmarkEchoGitHubAPI-8     	   24586	     47607 ns/op	    3032 B/op	     203 allocs/op
BenchmarkEchoGplusAPI-8      	  486565	      2554 ns/op	     175 B/op	      13 allocs/op
BenchmarkEchoParseAPI-8      	  251629	      4572 ns/op	     345 B/op	      26 allocs/op
BenchmarkGinStatic-8         	   34180	     33394 ns/op	    8550 B/op	     157 allocs/op
BenchmarkGinGitHubAPI-8      	   21962	     48168 ns/op	   11323 B/op	     203 allocs/op
BenchmarkGinGplusAPI-8       	  458335	      2761 ns/op	     699 B/op	      13 allocs/op
BenchmarkGinParseAPI-8       	  232579	      5035 ns/op	    1397 B/op	      26 allocs/op
BenchmarkForestStatic-8      	   37356	     32453 ns/op	    2183 B/op	     157 allocs/op
BenchmarkForestGitHubAPI-8   	   24778	     45812 ns/op	    3530 B/op	     188 allocs/op
BenchmarkForestGplusAPI-8    	  477079	      2469 ns/op	     176 B/op	      13 allocs/op
BenchmarkForestParseAPI-8    	  254059	      4672 ns/op	     588 B/op	      22 allocs/op
PASS
ok  	github.com/honmaple/forest/examples/benchmark	17.770s
```
