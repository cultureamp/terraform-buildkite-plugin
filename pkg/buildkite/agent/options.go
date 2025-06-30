package agent

type AnnotationStyle string

const (
	StyleSuccess AnnotationStyle = "success"
	StyleInfo    AnnotationStyle = "info"
	StyleWarning AnnotationStyle = "warning"
	StyleError   AnnotationStyle = "error"
)

type annotateConfig struct {
	message  string
	style    AnnotationStyle
	context  string
	artifact string
	append   bool
}

type AnnotateOptions func(*annotateConfig)

func WithMessage(m string) AnnotateOptions {
	return func(r *annotateConfig) {
		r.message = m
	}
}

func WithStyle(s AnnotationStyle) AnnotateOptions {
	return func(r *annotateConfig) {
		r.style = s
	}
}

func WithContext(c string) AnnotateOptions {
	return func(r *annotateConfig) {
		r.context = c
	}
}

func WithArtifact(a string) AnnotateOptions {
	return func(r *annotateConfig) {
		r.artifact = a
	}
}

func WithAppend(a bool) AnnotateOptions {
	return func(r *annotateConfig) {
		r.append = a
	}
}
