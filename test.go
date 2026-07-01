package main

import (
	"fmt"
	"github.com/kmmz532/php2js/internal/transformer"
)

func main() {
	fmt.Println("key reserved?", transformer.IsReserved("key"))
}
