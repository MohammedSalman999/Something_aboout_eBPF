package main

import "fmt"

func main() {
    // Create a buffered channel 'cnp' that can hold up to 10 functions.
    cnp := make(chan func(), 10)
    
    // Launch 4 goroutines, each handling functions received from 'cnp'.
    for i := 0; i < 4; i++ {
        // Each goroutine starts anonymously with 'go func() {...}()'.
        go func() {
            // Continuously receive and execute functions from 'cnp'.
            for f := range cnp {
                f()  // Execute the received function.
            }
        }()
    }
    
    // Send a function to 'cnp' that prints "HERE1".
    cnp <- func() {
        fmt.Println("HERE1")
    }
    
    // Print "Hello" to the console.
    fmt.Println("Hello")
}





