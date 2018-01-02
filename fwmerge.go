package main

import (
	"fmt"
	flag "github.com/ogier/pflag"
	"os"
	"strings"
)

const USAGE_TEXT = `
fwmerge - a firewall ruleset renderer [v0.1]

fwmerge is a firewall ruleset renderer for firewalls that support a table/chain/rule
structure(like iptables and nftables). It takes YAML files as inputs and outputs a
ruleset that can be loaded into the given firewall. Each rule is tagged with a priority
allowing fwmerge to merge chains and sort the rules. The final ruleset is then output
in the requested format. fwmerge doesn't know about specific rules, and cannot translate
between different firewall syntaxes.

A sample rule file YAML could look like the following:
  filter:
    INPUT:
      - policy: DROP
      - 10 allow ICMP: -p icmp -j ACCEPT
      - 10 allow all on loopback: -i lo -j ACCEPT
      - 10 allow SSH: -p tcp --dport 22 -j ACCEPT
    testchain: unmanaged

fwmerge can set the policy for chains that support it using the policy tag. Note that
the policy tag has no prioirty, the last policy set will win.

fwmerge can also create unmanaged chains. These are chains that fwmerge will ask the
firewall to create, but it won't output rules to populate the chain. This allows other
applications to manage these chains without interference.

The rules are specified as either:
  <priority>: <rule>
  <priority> <comment>: <rule>

The priority is used for sorting, the comment is ignored, and the rule is output
verbatim into the ruleset. The rule must be convertable to a string.

For now the only supported generator is iptables. This generator will output a ruleset
that can be piped to iptables-restore.
`

func main() {
	generator := flag.String("generator", "iptables", "Which generator to use.")
	flag.SetInterspersed(false)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, strings.TrimSpace(USAGE_TEXT)+"\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file...>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		flag.Usage()
		return
	}

	rs, err := ParseFiles(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing rulesets:\n%v\n", err)
		return
	}

	switch *generator {
	case "iptables":
		ipt, err := GenerateIptables(rs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating iptables output: %v\n", err)
			return
		}
		fmt.Print(ipt)
	default:
		fmt.Fprintf(os.Stderr, "Unknown generator '%s'", *generator)
	}
}