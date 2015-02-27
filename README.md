# jsonpath  
  
jsonpath is used to pull values out of a JSON document without unmarshalling the string into an object.  At the loss of post-parse random access and conversion to primitive types, you gain faster return speeds and lower memory utilization.  If the value you want is located near the start of the json, the evaluator will terminate after reaching and recording its destination.  
  
The evaluator can be initialized with several paths, so you can retrieve multiple sections of the document with just one scan.  Naturally, when all paths have been reached, the evaluator will early terminate.  
  
For each value returned by a path, you'll also get the keys & indexes needed to reach that value.  Use the `keys` flag to view this in the CLI.  The Go package will return an `[]interface{}` of length `n` with indexes `0 - (n-2)` being the keys and the value at index `n-1`.  
  
### CLI   
`go get github.com/NodePrime/jsonpath/cli/jsonpath`  
`jsonpath [-file="FILEPATH"] [-json="{...}"] [-keys] -path='PATH'` 

##### Usage  
  `-f, --file="": Path to json file  
  -j, --json="": JSON text  
-k, --keys=false: Print keys & indexes that lead to value  
-p, --path=[]: One or more paths to target in JSON`  

  
### Go Package  
go get github.com/NodePrime/jsonpath  
  
`eval, err := jsonpath.GetPathsInBytes(json []byte, pathStrings ...string) (*jsonpath.eval, error)`  
or  
`eval, err := jsonpath.GetPathsInReader(r io.Reader, pathStrings ...string) (*jsonpath.eval, error)`  
  
then
```go
for r := range eval.Results {
	// skip keys/indexes & get final byte data
	fmt.Println(r[len(r)-1].([]byte]))	
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
*CLI*  
`jsonpath -file=example.json -path='$.Items[*].tags[*]+' -keys`  
Items	0	tags	0	"comedy"  
Items	0	tags	1	"shakespeare"  
Items	0	tags	2	"play"  
Items	1	tags	0	"french"  
Items	1	tags	1	"revolution"  
Items	1	tags	2	"london"  

*Package*  
`$.Items[*].title+`   
... "A Midsummer Night's Dream"   
... "A Tale of Two Cities"   
  
`$.Items[*].tags+`    
... ["comedy","shakespeare","play"]  
... ["french","revolution","london"]  
  
`$.Items[*].tags[*]+`  
... "comedy"  
... "shakespeare"  
... "play"  
... "french"  
... "revolution"  
...  "london"  
  
... = keys/indexes of path  
  
  
### TODO  
*Expressions*  
`?(expression)`  
Ex: `$.Items[*]?(title=='A Tale of Two Cities').tags+`  
  
*Efficient Reader Lexer*  
Paths tell lexer whether it should be appending bytes to the lexeme buffer  
(might not be a huge benefit)  
