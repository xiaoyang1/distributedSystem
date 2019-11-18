package mapreduce

import (
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"log"
	"os"
	"sort"
)

func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTask int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	//
	// doReduce manages one reduce task: it should read the intermediate
	// files for the task, sort the intermediate key/value pairs by key,
	// call the user-defined reduce function (reduceF) for each key, and
	// write reduceF's output to disk.
	//
	// You'll need to read one intermediate file from each map task;
	// reduceName(jobName, m, reduceTask) yields the file
	// name from map task m.
	//
	// Your doMap() encoded the key/value pairs in the intermediate
	// files, so you will need to decode them. If you used JSON, you can
	// read and decode by creating a decoder and repeatedly calling
	// .Decode(&kv) on it until it returns an error.
	//
	// You may find the first example in the golang sort package
	// documentation useful.
	//
	// reduceF() is the application's reduce function. You should
	// call it once per distinct key, with a slice of all the values
	// for that key. reduceF() returns the reduced value for that key.
	//
	// You should write the reduce output as JSON encoded KeyValue
	// objects to the file named outFile. We require you to use JSON
	// because that is what the merger than combines the output
	// from all the reduce tasks expects. There is nothing special about
	// JSON -- it is just the marshalling format we chose to use. Your
	// output code will look something like this:
	//
	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
	// Your code here (Part I).
	//

	keyValues := make(map[string][]string)
	// 从多个map中读取文件
	for mapTask := 0; mapTask < nMap; mapTask++ {
		fileName := reduceName(jobName, mapTask, reduceTask)
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatal(err)
		}

		decoder := json.NewDecoder(file)

		for decoder.More() {
			var kv KeyValue
			err := decoder.Decode(&kv)
			if err != nil {
				log.Fatal(err)
			}
			keyValues[kv.Key] = append(keyValues[kv.Key], kv.Value)
		}

		file.Close()
	}

	// 对所有的key排序
	var keys []string
	for k := range keyValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 写入临时结果文件
	_uuid, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	tmpFile := outFile + "-uuid-" + _uuid.String()
	out, err := os.Create(tmpFile)
	if err != nil {
		log.Fatal(err)
	}

	// 将临时文件转化为最终文件
	defer func() {
		out.Close()
		if !FileExist(outFile) {
			os.Rename(tmpFile, outFile)
		}
	}()

	encoder := json.NewEncoder(out)
	for _, key := range keys {
		result := reduceF(key, keyValues[key])
		err = encoder.Encode(KeyValue{key, result})
		if err != nil {
			log.Fatal("failed to encode", KeyValue{key, result})
		}
	}
}
