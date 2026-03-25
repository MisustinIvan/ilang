; Keywords
["return" "let" "extrn" "if" "else" "for" "make" "release"] @keyword

; Built-in Types
(basic_type) @type
(pointer_type "^" @type.punctuation)
(array_type ["[" "]" ] @punctuation.bracket)

; Functions
(declaration name: (identifier) @function)
(call (identifier) @function.call)

; Literals
(int_literal) @number
(float_literal) @number.float
(string_literal) @string
(bool_literal) @boolean
(comment) @comment

; Operators & Punctuators
(binary_operator) @operator
(unary_operator) @operator
["(" ")" "[" "]" "{" "}"] @punctuation.bracket
["," ";" ":" "="] @punctuation.delimiter

; Variables
(identifier) @variable
(function_argument (identifier) @variable.parameter)
(bind (identifier) @variable.declaration)
