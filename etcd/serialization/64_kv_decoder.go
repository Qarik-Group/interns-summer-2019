package main

import (
	"bufio"
  	// "flag"
  	"fmt"
	"strings"
  	"io"
 	 "os"
	// "context"
	// "log"
	// "time"
	// "go.etcd.io/etcd/clientv3"
	"encoding/base64"
)

func main() {

//open the file
// file, err := os.OpenFile("encoded.txt", os.O_RDONLY, 0644)
         // if err != nil {
         //         panic(err)
         // }

         // defer file.Close()



         reader := bufio.NewReader(os.Stdin)
				 i:=0
         for {
					 	i+=1

          	line, _, err := reader.ReadLine()
						if err == io.EOF {
                 break
            }

						a:= strings.Split(string(line), " : ")
						str_key := a[0]
						str_val := a[1]

						data_key, err := base64.StdEncoding.DecodeString(str_key)
						data_val, err := base64.StdEncoding.DecodeString(str_val)
						if err != nil {
							fmt.Println("error:", err)
							return
						}
						// fmt.Printf("%s\n", data_key)
						// fmt.Printf("%s\n", data_val)

						//output to a file instead of stdout
						// f, err := os.OpenFile("decoded.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						// if err != nil {
						// 		fmt.Println(err)
						// 		return
						// }



						//format to k:v
						newLine := fmt.Sprintf("%s : %s", data_key, data_val)
						fmt.Println(newLine)
				    //append the kvs
				    // _, err = fmt.Fprintln(f, newLine)
				    // if err != nil {
				    //     fmt.Println(err)
				    //             f.Close()
				    //     return
				    // }

						//close the file
						// err = f.Close()
						// if err != nil {
						// 		fmt.Println(err)
						// 		return
						// }

         }

//close the file
// err = file.Close()
// if err != nil {
// 		fmt.Println(err)
// 		return
// }

// fmt.Println(i,"lines read successfully")
}
