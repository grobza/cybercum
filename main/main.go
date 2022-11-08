package main

import (
    cum "github.com/dyvdev/cybercum"
    "flag"
)

func main() {
    read := flag.Bool("read", false, "read")
    flag.Parse()

    if *read {
        cum.ReadBot("./config.json")
    } else {
        cum.RunBot("./config.json")
    }
}
