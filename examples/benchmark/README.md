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
BenchmarkEchoStatic-8        	   36494	     32263 ns/op	    2205 B/op	     157 allocs/op
BenchmarkEchoGitHubAPI-8     	   24674	     49399 ns/op	    3027 B/op	     203 allocs/op
BenchmarkEchoGplusAPI-8      	  463706	      2615 ns/op	     178 B/op	      13 allocs/op
BenchmarkEchoParseAPI-8      	  257562	      4471 ns/op	     342 B/op	      26 allocs/op
BenchmarkGinStatic-8         	   33537	     33098 ns/op	    8570 B/op	     157 allocs/op
BenchmarkGinGitHubAPI-8      	   23739	     47731 ns/op	   11204 B/op	     203 allocs/op
BenchmarkGinGplusAPI-8       	  430006	      2655 ns/op	     704 B/op	      13 allocs/op
BenchmarkGinParseAPI-8       	  220128	      4984 ns/op	    1405 B/op	      26 allocs/op
BenchmarkForestStatic-8      	   31466	     36726 ns/op	    2356 B/op	     157 allocs/op
BenchmarkForestGitHubAPI-8   	   19384	     61364 ns/op	    4094 B/op	     188 allocs/op
BenchmarkForestGplusAPI-8    	  408818	      2923 ns/op	     188 B/op	      13 allocs/op
BenchmarkForestParseAPI-8    	  224467	      5096 ns/op	     642 B/op	      22 allocs/op
PASS
ok  	github.com/honmaple/forest/examples/benchmark	17.770s
```
