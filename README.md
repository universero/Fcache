# FCache

A simple distributed cache system for learning in golang

> thanks for [geektutu](https://geektutu.com/post/geecache.html)

the document only record some personally important part for me, and the change of code step by step will be hard to
find, so it's highly recommend to view the original blog.

```text
├─code: related codes
|
└─README.md: related documents
```

## Cache Obsolescence Strategy

- FIFO
  The strategy FIFO(First in First out) means always obsolete the oldest record,
  based on the assumption that the oldest record is the least likely to be used again.
  To release the FIFO, we only need a queue. However, in some cases, some records which are
  added firstly also are the most frequently used, but they have to be added and discarded frequent because of the FIFO
  strategy
- LFU
  The strategy LFU(Least Frequently Used) tend to discard the record which has the least frequency of use,
  based on the assumption that the most used one is more likely to be retrieved.
  To release this strategy, we need to maintain a queue ordered by visit times. This strategy has a high hit ratio, but
  also has an obvious defect.
  It's costly to maintain the number of each visit, and some record that is visited frequently in the queue may never be
  visited again, but will not be removed with a high number of visit
- LRU
  The strategy LRU(Least Recently Used) is a trade off between the time factor and retrieve frequency. It assumes that a
  record visited recently will be visited again in the future. To release it,
  we need a queue to restore the records, and each time we visit a record, move it to the tail, therefore the head of
  the queue is the least recently used one, if we need to remove one, discard it.
  **in the code the LRU one is at back**

## sync.Mutex

When more than one goroutine do I/O on a variable at same time, some conflicts may happen.
To ensure only one goroutine can access the variable at once, which called mutual exclusion, we need the lib sync.Mutex.

Here is a simple case

```go
package main

import (
	"fmt"
	"time"
)

var set = make(map[int]bool, 0)

func printOnce(num int) {
	if _, exist := set[num]; !exist {
		fmt.Println(num)
	}
	set[num] = true
}

func main() {
	for i := 0; i < 10; i++ {
		go printOnce(100)
	}
	time.Sleep(time.Second)
}
```

Try to run `go run .`, and you will find that the number of the programme print are different.
Sometimes it prints twice, sometimes only once, and even it may print 10 times.

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

var m sync.Mutex
var set = make(map[int]bool, 0)

func printOnce(num int) {
	m.Lock()
	if _, exist := set[num]; !exist {
		fmt.Println(num)
	}
	set[num] = true
	m.Unlock()
}

func main() {
	for i := 0; i < 10; i++ {
		go printOnce(100)
	}
	time.Sleep(time.Second)
}
```

Using `Lock()` and `UnLock()` to wrap the access to the set, and there will be no conflict
The programme will print only once

## Group

                                     Y

accept key --> check whether cache -----> return the cache value ⑴
| N Y
|-----> whether to get from the remote -----> 与远程节点交互 --> return the value ⑵
| N
|-----> call `callback function`，get the value and add it to the cache --> return the value

Group is the main struct to interact with users , store the cache value and get cache value from different way;

## HTTP Server

Use HTTP Server to get data from http protocol in http.go

## Consistent Hash

- Which node to visit ?

For distributed cache, when a node receive a request for a key which will miss, it's hard to decide which node to load
the missing value;
Assuming that there are totally 10 nodes, if we choose a node called 1 randomly, node 1 load the key from the data
source;
and the second time, if we still choose randomly, there is only tenth to choose node 1, and nine tenth to choose other
node, which means load the key from the data source again and costs a lot.

How about divide the sum of key's ASCII Code by 10, and choose the node with the reminder?
It looks like feasible. However, if one node is down, almost every key cannot be found in the cache.
When lots of cache fail at a same time, DB will receive a large mount of requests and have huge pressure suddenly.
When it happens, we call it a cache avalanche.

- Consistent Hash

To solve the problem, we can use the consistent hash.
In the algorithm, we map the node to 0~2^32 with its name or ip etc. , and organize the numbers like a ring.
When we get a key, calculate its hash and find the first node on the ring clockwise.
If the number of nodes changed, we only need to reorganize a few of key.

- Virtual Node

If there is not enough nodes, it may happen that a node holds most keys, and others have a few of them.
To solve it, we can use virtual node(a real nodes correspond to multiple virtual node) to balance the load of different
nodes.

## Get cache from Distribute Node

When a key missed, cache will try to load it from data source before, now in this part (code/4), it will try to load
from other distribute node first;

## single flight

If a key is required concurrently, there will be lost of requests send to other peer to find the same key. However, in
fact, we only need to send only one
request to sever all the key required. So we use the single flight to do that multiple access to a key in a short time
and only one request to other peer;

## TODO

Now only the response use the proto, but the request is still based on the http.
May can change it to pure proto and rpc;
Also after learning distribute system, maybe can optimize and improve the whole project;