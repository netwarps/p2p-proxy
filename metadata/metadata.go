package metadata

import "fmt"

var Version = "v0.0.1"

var CommitSHA = "commit"

var Banner = `

   ██████╗ ██████╗ ██████╗       ██████╗ ██████╗  ██████╗ ██╗  ██╗██╗   ██╗
   ██╔══██╗╚════██╗██╔══██╗      ██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝╚██╗ ██╔╝
   ██████╔╝ █████╔╝██████╔╝█████╗██████╔╝██████╔╝██║   ██║ ╚███╔╝  ╚████╔╝
   ██╔═══╝ ██╔═══╝ ██╔═══╝ ╚════╝██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗   ╚██╔╝
   ██║     ███████╗██║           ██║     ██║  ██║╚██████╔╝██╔╝ ██╗   ██║
   ╚═╝     ╚══════╝╚═╝           ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝

                     A http(s) proxy based on P2P

                Github: https://github.com/diandianl/p2p-proxy
                Version: %s
                CommitSHA: %s

`

func PrintBanner() {
	if len(Banner) > 0 {
		fmt.Printf(Banner, Version, CommitSHA)
	}
}
