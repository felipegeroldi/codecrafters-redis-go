package main

const (
	REDIS_STR      = byte('+')
	REDIS_ERR      = byte('-')
	REDIS_INT      = byte(':')
	REDIS_ARR      = byte('*')
	REDIS_NIL      = byte('_')
	REDIS_MAPS     = byte('%')
	REDIS_SETS     = byte('~')
	REDIS_PUSH     = byte('>')
	REDIS_BOOL     = byte('#')
	REDIS_DOUBLE   = byte(',')
	REDIS_BIGNUM   = byte('(')
	REDIS_BULK_ERR = byte('!')
	REDIS_VERB_STR = byte('=')
	REDIS_BULK_STR = byte('$')
)
