package main

import (
	"fmt"
	"log"
	"os"
)

// Creates a very deep directory directory structure that
// exhausts system memory
func main() {
	path := "/tmp/deep"
	os.Mkdir(path, 0755)
	os.Chdir(path)

	// 2^30 is big enough that the process will run out of memory before completing on most machines
	// 2^10 (1024) is a good value to use for profiling memory usage
	// n := 10
	n := 30
	for i := 0; i < (1 << n); i++ {
		// dirname := strings.Repeat("x", 255)
		dirname := "x"

		// print depth
		if i%128 == 0 {
			fmt.Println(i)
		}

		err := os.Mkdir(dirname, 0755)
		if err != nil {
			log.Fatal(err)
		}
		os.Chdir(dirname)
	}
}
