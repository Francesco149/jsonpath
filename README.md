# jsonpath  
  
jsonpath is used to pull values out of a JSON document without unmarshalling the string into an object.  At the loss of post-parse random access and conversion to primitive types, you gain faster return speeds and lower memory utilization.  If the value you want is located near the start of the json, the evaluator will terminate after reaching and recording its destination.  
  
The evaluator can be initialized with several paths, so you can retrieve multiple sections of the document with just one scan.  Naturally, when all paths have been reached, the evaluator will early terminate.  
  
For each value returned by a path, you'll also get the keys & indexes needed to reach that value.  Use the `showKeys` flag to view this in the CLI.  The Go package will return an `[]interface{}` of length `n` with indexes `0 - (n-2)` being the keys and the value at index `n-1`.  
  
### CLI   
`go get github.com/NodePrime/jsonpath/cli/jsonpath`  
`jsonpath [-file="FILEPATH"] [-json="{...}"] [-showKeys] -path='PATH'` 
  
You can specify more than one path by repeating the `path` flag.  If you do not use the `-file` or `-json` flags, then you can pipe JSON to StdIn.  
  
### Go Package  
go get github.com/NodePrime/jsonpath  
  
`eval, err := jsonpath.GetPathsInBytes(json []byte, pathStrings ...string) (*jsonpath.eval, error)`  
or  
`eval, err := jsonpath.GetPathsInReader(r io.Reader, pathStrings ...string) (*jsonpath.eval, error)`  
  
then
```go
for r := range eval.Results {
	// skip keys/indexes & print value string
	fmt.Println(r[len(r)-1].(string))	
}
if eval.Error != nil {
	return eval.Error
}
```
^ this interface may change   
  
   
### Path Syntax  
All paths start from the root node `$`.  Similar to getting properties in a JavaScript object, a period `.` or brackets `[ .. ]` are used.  
  
`$` = root  
`.` = property of  
`["abc"]` = quoted property name  
`*` = wildcard property name  
`[n]` = Nth index of array  
`[n:m]` = Nth index to m-1 index (same as Go Slicing)  
`[n:]` = Nth index to end of array  
`[*]` = wildcard index of array  
`+` = get value at end of path  
  
Example: 
```javascript
{  
	"Items":   
		[  
			{  
				"title": "A Midsummer Night's Dream",  
				"tags":[  
					"comedy",  
					"shakespeare",  
					"play"  
				]  
			},{  
				"title": "A Tale of Two Cities",  
				"tags":[  
					"french",  
					"revolution",  
					"london"  
				]  
			}  
		]  
} 
```
	
Example Paths:   
`$.Items[*].title+`    
"A Midsummer Night's Dream"   
"A Tale of Two Cities"   
  
`$.Items[*].tags+`    
["comedy","shakespeare","play"]  
["french","revolution","london"]  
  
`$.Items[*].tags[*]+`  
"comedy"  
"shakespeare"  
"play"  
"french"  
"revolution"  
"london"  
