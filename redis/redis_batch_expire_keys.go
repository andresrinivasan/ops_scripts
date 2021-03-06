package main

// MAKE SURE TO SET export GOMAXPROCS=16  in your environment before running

import "fmt"
import "flag"
import "time"
import "os"
import "github.com/fzzy/radix/redis"

func fetchAllKeys(hostame string, port int, database int) []string {
	c, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostname, port), time.Duration(3000)*time.Second)
	errHndlr(err)
	//keys := c.Cmd("SELECT", database)
	keys := c.Cmd("KEYS", "URLPARSER_*")
	fmt.Println("got keys")
	j := keys.Elems
	fmt.Println("keys count ", len(j))
	redis_keys := make([]string, len(j), len(j))
	for i := 0; i < len(j); i++ {
		redis_keys[i] = fmt.Sprintf("%s", j[i])
	}
	return redis_keys
}

func errHndlr(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func worker(id int, jobs <-chan string, results chan<- string, hostame string, port int, database int) {
	c, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostname, port), time.Duration(3)*time.Second)
	errHndlr(err)
	for j := range jobs {
		errHndlr(err)
		k := c.Cmd("SELECT", database)
		k = c.Cmd("EXPIRE", j, "2")
		fmt.Println("expring ", j)
		if 3 < 2 {
			fmt.Println(k.Elems[0].Str())
		}
		results <- "OK"
	}
	c.Close()
}

var hostname string
var port int
var batch_size int
var concurrent int
var database int

func init() {
	flag.StringVar(&hostname, "hostname", "localhost", "hostname or ip to scan")
	flag.IntVar(&port, "port", 6379, "port to try to connect to")
	flag.IntVar(&concurrent, "concurrent", 10, "number of workers to run")
	flag.IntVar(&database, "database", 0, "Redis database to use.  DB 0 is the default")
	flag.IntVar(&batch_size, "batch-size", 10000, "Batch size to use default is 10K")
	flag.Parse()
}

func main() {

	keys := fetchAllKeys(hostname, port, database)
	fmt.Println("Expiring ", batch_size, " keys out of ", len(keys))
	if len(keys) < batch_size {
		fmt.Println("Batch size is greater than the number of keys")
		os.Exit(1)
	}
	// In order to use our pool of workers we need to send
	// them work and collect their results. We make 2
	// channels for this.
	jobs := make(chan string, batch_size)
	results := make(chan string, batch_size)

	for w := 0; w <= concurrent; w++ {
		go worker(w, jobs, results, hostname, port, database)
	}

	for j := 0; j <= batch_size-1; j++ {
		jobs <- keys[j]
	}
	close(jobs)

	// Finally we collect all the results of the work.
	for a := 0; a <= batch_size-1; a++ {
		<-results
	}
	os.Exit(0)
}
