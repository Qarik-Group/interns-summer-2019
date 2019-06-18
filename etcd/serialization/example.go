package main

import (
	"context"
  "fmt"
	"log"
	"time"
	"go.etcd.io/etcd/clientv3"
  "os"
)

func main() {
cli, err := clientv3.New(clientv3.Config{
  Endpoints:   []string{"http://10.129.0.11:31497"},
  DialTimeout: 2 * time.Second,
})
if err != nil {
    log.Fatal(err)
}
defer cli.Close()

// for i := range make([]int, 3) {
//     ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//     _, err = cli.Put(ctx, fmt.Sprintf("key_%d", i), "value")
//     cancel()
//     if err != nil {
//         log.Fatal(err)
//     }
// }

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
resp, err := cli.Get(ctx, "", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
cancel()
if err != nil {
    log.Fatal(err)
}

i:=0

for _, ev := range resp.Kvs {
    i +=1
    //printing kv to stdout
    // fmt.Printf("%s : %s\n", ev.Key, ev.Value)

    //open the file
    f, err := os.OpenFile("result.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println(err)
        return
    }

    //get the kvs
    newLine := fmt.Sprintf("%s : %s", ev.Key, ev.Value)

    //append the kvs
    _, err = fmt.Fprintln(f, newLine)
    if err != nil {
        fmt.Println(err)
                f.Close()
        return
    }
    //close the file
    err = f.Close()
    if err != nil {
        fmt.Println(err)
        return
    }

}

fmt.Println(i,"lines appended successfully")
}
