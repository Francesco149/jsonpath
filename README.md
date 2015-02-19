# jsonpath  
  
jsonpath is built to pull values out of a JSON document without unmarshalling the entire string.  You can specify one or more paths to specify what sections you'd like returned.   
  
### CLI   
go get github.com/NodePrime/jsonpath/cli/jsonpath  TODO: Move this  
jsonpath [-file="FILEPATH"] [-json="JSON"] -path='PATH' 
  
You can specify more than one path by repeating the path flag.  
  
TODO: Std-in json  
  
### Go Package  
go get github.com/NodePrime/jsonpath  
`jsonpath.GetPathsInBytes(json []byte, pathStrings ...string) (*jsonpath.eval, error)`  
`jsonpath.GetPathsInReader(r io.Reader, pathStrings ...string) (*jsonpath.eval, error)`  
   
### Path Syntax  
All paths start from the root node `$`.  Similar to getting properties in a JavaScript object, a period `.` or brackets `[ .. ]` are used.  
  
$ = root  
. = property of  
[**] = wildcard index of array  
[N] = Nth index of array  
** = wildcard property name  
"P" = quoted property name in case property contains `.` or `[`  
  
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
`$.Items[*].title`    
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
