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
BenchmarkEchoStatic-8        	   30968	     32932 ns/op	    2374 B/op	     157 allocs/op
BenchmarkEchoGitHubAPI-8     	   24879	     49490 ns/op	    3016 B/op	     203 allocs/op
BenchmarkEchoGplusAPI-8      	  445206	      2623 ns/op	     181 B/op	      13 allocs/op
BenchmarkEchoParseAPI-8      	  256200	      4766 ns/op	     343 B/op	      26 allocs/op
BenchmarkGinStatic-8         	   26940	     43482 ns/op	   11949 B/op	     314 allocs/op
BenchmarkGinGitHubAPI-8      	   19398	     60057 ns/op	   15512 B/op	     406 allocs/op
BenchmarkGinGplusAPI-8       	  350064	      3465 ns/op	    1035 B/op	      26 allocs/op
BenchmarkGinParseAPI-8       	  176620	      6509 ns/op	    2068 B/op	      52 allocs/op
BenchmarkForestStatic-8      	   24504	     50018 ns/op	    1963 B/op	     157 allocs/op
BenchmarkForestGitHubAPI-8   	   12926	     91675 ns/op	    9479 B/op	     287 allocs/op
BenchmarkForestGplusAPI-8    	  269132	      4288 ns/op	     455 B/op	      17 allocs/op
BenchmarkForestParseAPI-8    	  137349	      8774 ns/op	    1833 B/op	      44 allocs/op
PASS
ok  	github.com/honmaple/forest/examples/benchmark	18.706s
```
