module.exports = grammar({
	name: 'ilang',

	extras: $ => [
		/\s/,
		$.comment,
	],

	rules: {
		program: $ => repeat(choice(
			$.declaration,
			$.external_declaration,
		)),

		comment: $ => seq('#', /.*/),

		declaration: $ => seq(
			$.basic_type,
			field('name', $.identifier),
			'(',
			optional(seq($.function_argument, repeat(seq(',', $.function_argument)))),
			')',
			$.block
		),

		external_declaration: $ => seq(
			'extrn',
			$.basic_type,
			$.identifier,
			'(',
			choice(
				seq(optional(seq($.function_argument, repeat(seq(',', $.function_argument)))), optional(seq(',', '...'))),
				'...'
			),
			')'
		),

		function_argument: $ => seq($.type, $.identifier),

		block: $ => seq(
			'{',
			repeat(seq($.expression, ';')),
			optional($.expression),
			'}'
		),

		expression: $ => choice(
			$.return,
			$.bind,
			$.assignment,
			$.value
		),

		value: $ => choice(
			$.primary,
			$.binary,
			$.unary
		),

		return: $ => seq('return', $.value),
		bind: $ => seq('let', $.identifier, ':', $.type, '=', $.value),

		assignment: $ => seq(
			choice($.identifier, $.index, $.deref),
			'=',
			$.value
		),

		deref: $ => prec(1, seq('@', $.identifier)),

		binary: $ => prec(1, prec.left(seq($.primary, $.binary_operator, $.value))),

		unary: $ => prec(2, seq($.unary_operator, $.primary)),

		index: $ => prec(1, seq($.identifier, '[', $.primary, ']')),

		primary: $ => choice(
			$.literal,
			$.identifier,
			$.call,
			$.separated,
			$.block,
			$.condition,
			$.index,
			$.deref,
			$.loop,
			$.make,
			$.release
		),

		make: $ => seq('make', '(', $.basic_type, ',', $.value, ')'),
		release: $ => seq('release', '(', $.identifier, ')'),
		loop: $ => seq('for', $.value, $.block),
		call: $ => prec(1, seq($.identifier, '(', optional(seq($.value, repeat(seq(',', $.value)))), ')')),
		separated: $ => seq('(', $.value, ')'),

		condition: $ => prec.right(seq(
			'if', $.value, $.value,
			optional(seq('else', $.value))
		)),

		type: $ => choice($.basic_type, $.array_type, $.slice_type, $.pointer_type),
		basic_type: $ => choice('int', 'bool', 'float', 'string', 'unit'),
		array_type: $ => seq('[', $.int_literal, ']', $.basic_type),
		slice_type: $ => seq('[', optional($.identifier), ']', $.basic_type),
		pointer_type: $ => seq('^', $.basic_type),

		literal: $ => choice($.int_literal, $.float_literal, $.string_literal, $.bool_literal, $.array_literal),
		int_literal: $ => /\d+/,
		float_literal: $ => /\d+\.\d*/,
		string_literal: $ => /"[^"]*"/,
		bool_literal: $ => choice('true', 'false'),
		array_literal: $ => seq('[', optional(seq($.value, repeat(seq(',', $.value)))), ']'),

		identifier: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,

		binary_operator: $ => choice('+', '-', '*', '/', '==', '!=', '<', '>', '<=', '>=', '<<', '>>', '&&', '||'),
		unary_operator: $ => choice('-', '!', '^', '@'),
	}
});
