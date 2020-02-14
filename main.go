/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"fmt"

	"github.com/diandianl/p2p-proxy/cmd"
	"github.com/diandianl/p2p-proxy/signal"

)

const logo = `

   ██████╗ ██████╗ ██████╗       ██████╗ ██████╗  ██████╗ ██╗  ██╗██╗   ██╗
   ██╔══██╗╚════██╗██╔══██╗      ██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝╚██╗ ██╔╝
   ██████╔╝ █████╔╝██████╔╝█████╗██████╔╝██████╔╝██║   ██║ ╚███╔╝  ╚████╔╝ 
   ██╔═══╝ ██╔═══╝ ██╔═══╝ ╚════╝██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗   ╚██╔╝  
   ██║     ███████╗██║           ██║     ██║  ██║╚██████╔╝██╔╝ ██╗   ██║   
   ╚═╝     ╚══════╝╚═╝           ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   

               v0.0.1    A http(s) proxy based on P2P

`

func main() {

	fmt.Println(logo)

	ctx := context.Background()

	done, ctx := signal.SetupInterruptHandler(ctx)

	defer done()

	cmd.Execute(ctx)
}
