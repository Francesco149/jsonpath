Path Syntax Elements  

$ - root object/element  
@ - current object  
. or []  - child operator  
- parent operator  
* wildcard. All objets/elements regardless their names  
[] - native array operator  
[:] - slice functionality  
| - OR operator Branch path with value set  
.length - returns length of array  



Use cases  
Pull out values  
Pull out keys  
Pull out slice of array  

Proposed Output  
[][]string - Dynamic matrix of strings  

Ex:  
{ "store": {  
    "book": [   
      { "category": "reference",  
        "author": "Nigel Rees",  
        "title": "Sayings of the Century",  
        "price": 8.95  
      },  
      { "category": "fiction",  
        "author": "Evelyn Waugh",  
        "title": "Sword of Honour",  
        "price": 12.99  
      },  
      { "category": "fiction",  
        "author": "Herman Melville",  
        "title": "Moby Dick",  
        "isbn": "0-553-21311-3",  
        "price": 8.99  
      },  
      { "category": "fiction",  
        "author": "J. R. R. Tolkien",  
        "title": "The Lord of the Rings",  
        "isbn": "0-395-19395-8",  
        "price": 22.99  
      }
    ],
    "bicycle": [{  
      "color": "red",  
      "price": 19.95  
    }]  
  }  
}  
  
$.store.*[*].price|isbn    
Results:  
For each key, range, or evaluatable key/index, a column with that value will be populated. The final column will be the value  
  
[  
	["store", "book", 0, 8.95],  
	["store", "book", 1, 12.99],  
	["store", "book", 2, 8.99],  
	["store", "bicycle", 0, 19.95]  
]  







