package main

import (
	"context"
  "fmt"
	"log"
	"time"
	"go.etcd.io/etcd/clientv3"
  // "os"
	"encoding/base64"
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

// // for i := range make([]int, 3) {
//     ctxx, cancell := context.WithTimeout(context.Background(), 2*time.Second)
//
// 		// largedata:= "qweqweqweqweqwrijod fhsdhfiusgdfuiaugifaisgiwbeifbisdu ifuiegi%hwruohiashcdfuqbdj[];',./~!@#$%^&*(()*)_+nwjskcnjkbdxbnmbCbbzfmneanfqwbquwhoqhwriohqowrhoqhwroyq34y91265157013urehfhafslnzxcnqlrhoh183yriqfejasljfklnvkabsdfzbkbaekfgyuwegroqhaiodsjcjkbjkwashfdkljqihefewnlfsncjkbjksdjlxcndjs"
// 		largedata := "key:val%~!@#$%^&*()_+{}|:<>?,./;'[]-='\n"
// 	  _, err = cli.Put(ctxx, "key_large", fmt.Sprintf("value_%q", largedata))
//     cancell()
//     if err != nil {
//         log.Fatal(err)
//     }
// // }
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
resp, err := cli.Get(ctx, "", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
cancel()
if err != nil {
    log.Fatal(err)
}

//open the file
// f, err := os.OpenFile("encoded.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// if err != nil {
// 		fmt.Println(err)
// 		return
// }

i:=0
for _, ev := range resp.Kvs {
    i +=1
    //printing kv to stdout
    // fmt.Printf("%s : %s\n", ev.Key, ev.Value)


		//encode the key
		//encode the key
		data_k := []byte(ev.Key)
		data_v := []byte(ev.Value)
		encoded_k := base64.StdEncoding.EncodeToString(data_k)
		encoded_v := base64.StdEncoding.EncodeToString(data_v)
		// fmt.Println(str)

		//format to k:v
		newLine := fmt.Sprintf("%s : %s", encoded_k, encoded_v)
    //append the kvs
    // _, err = fmt.Fprintln(f, newLine)
		fmt.Println(newLine)
		// if err != nil {
    //     fmt.Println(err)
    //             f.Close()
    //     return
    // }
}
//close the file
// err = f.Close()
// if err != nil {
// 		fmt.Println(err)
// 		return
// }

// fmt.Println(i,"lines appended successfully")
}
