package errors

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"sort"
	"github.com/adyatlov/bun"
)

var errorsRegex = "error|failure|fail|critical"

func init() {
	builder := bun.CheckBuilder{
		Name: "errors-timeout-checker",
		Description: "Builds a timeline of failures to help find dependencies",
		CollectFromMasters:      collect,
		CollectFromAgents:       collect,
		CollectFromPublicAgents: collect,
		Aggregate:               aggregate,
	}
	check := builder.Build()
	bun.RegisterCheck(check)
}

func collect(host bun.Host) (ok bool, details interface{}, err error) {
	errorMatcher, _ := regexp.Compile(errorsRegex)
	keys := make(map[int]int)

	file, err1 := host.OpenFile("net")
	defer file.Close()

	ok = true
	if err1 != nil {
		ok = false
		errMsg := fmt.Sprintf("Cant open file %s", err1)
		fmt.Println(errMsg)
		return
	}

	scanner := bufio.NewScanner(file)
  for scanner.Scan() {
		line := scanner.Text()
		if errorMatcher.MatchString(line) {
			s := regexp.MustCompile(" ").Split(line, 3)
		  date := regexp.MustCompile(":").Split(s[1], 3)
			hours, _ := strconv.ParseFloat(date[0], 10)
			minutes, _ := strconv.ParseFloat(date[1], 10)
			minuteKey := int(hours*60 + (minutes/10))
			if _, ok := keys[minuteKey]; ok == false {
				keys[minuteKey] = 0
			}
			keys[minuteKey]++
		}
  }

  if err2 := scanner.Err(); err2 != nil {
		fmt.Println("Can't read file")
		ok = false
		return
  }

	fmt.Printf("Histogram %s\n", host.IP)
	sKeys := make([]int, 0, len(keys))
	for k := range keys {
		sKeys = append(sKeys, k)
	}
	sort.Ints(sKeys)

	for _, k := range sKeys {
		fmt.Printf("%2d: %d\n", k, keys[k])
	}
	fmt.Println("")

	details = keys
	ok = true
	return
}

func aggregate(c *bun.Check, b bun.CheckBuilder) {
	c.Summary = fmt.Sprintf("All versions are the same.")
}
