package vxml

import xmlimpl "github.com/imajinyun/go-knifer/internal/xml"

const (
	NBSP           = xmlimpl.NBSP
	AMP            = xmlimpl.AMP
	QUOTE          = xmlimpl.QUOTE
	APOS           = xmlimpl.APOS
	LT             = xmlimpl.LT
	GT             = xmlimpl.GT
	InvalidRegex   = xmlimpl.InvalidRegex
	CommentRegex   = xmlimpl.CommentRegex
	IndentDefault  = xmlimpl.IndentDefault
	ContentKey     = xmlimpl.ContentKey
	DefaultCharset = xmlimpl.DefaultCharset
)

type (
	Document        = xmlimpl.Document
	Element         = xmlimpl.Element
	Error           = xmlimpl.XMLError
	TokenHandler    = xmlimpl.TokenHandler
	NamespaceCache  = xmlimpl.NamespaceCache
	ParseOption     = xmlimpl.ParseOption
	WriteOption     = xmlimpl.WriteOption
	CleanOption     = xmlimpl.CleanOption
	BeanOption      = xmlimpl.BeanOption
	TransformOption = xmlimpl.TransformOption
	FormatOption    = xmlimpl.FormatOption
)
