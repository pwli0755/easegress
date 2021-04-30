package requestadaptor

import (
	"bytes"

	"github.com/megaease/easegateway/pkg/context"
	"github.com/megaease/easegateway/pkg/logger"
	"github.com/megaease/easegateway/pkg/object/httppipeline"
	"github.com/megaease/easegateway/pkg/supervisor"
	"github.com/megaease/easegateway/pkg/util/httpheader"
	"github.com/megaease/easegateway/pkg/util/pathadaptor"
	"github.com/megaease/easegateway/pkg/util/stringtool"
)

const (
	// Kind is the kind of RequestAdaptor.
	Kind = "RequestAdaptor"
)

var (
	results = []string{}
)

func init() {
	httppipeline.Register(&RequestAdaptor{})
}

type (
	// RequestAdaptor is filter RequestAdaptor.
	RequestAdaptor struct {
		super    *supervisor.Supervisor
		pipeSpec *httppipeline.FilterSpec
		spec     *Spec

		pa *pathadaptor.PathAdaptor
	}

	// Spec is HTTPAdaptor Spec.
	Spec struct {
		Method string                `yaml:"method" jsonschema:"omitempty,format=httpmethod"`
		Path   *pathadaptor.Spec     `yaml:"path,omitempty" jsonschema:"omitempty"`
		Header *httpheader.AdaptSpec `yaml:"header,omitempty" jsonschema:"omitempty"`
		Body   string                `yaml:"body" jsonschema:"omitempty"`
	}
)

// Kind returns the kind of RequestAdaptor.
func (ra *RequestAdaptor) Kind() string {
	return Kind
}

// DefaultSpec returns default spec of RequestAdaptor.
func (ra *RequestAdaptor) DefaultSpec() interface{} {
	return &Spec{}
}

// Description returns the description of RequestAdaptor.
func (ra *RequestAdaptor) Description() string {
	return "RequestAdaptor adapts request."
}

// Results returns the results of RequestAdaptor.
func (ra *RequestAdaptor) Results() []string {
	return results
}

// Init initializes RequestAdaptor.
func (ra *RequestAdaptor) Init(pipeSpec *httppipeline.FilterSpec, super *supervisor.Supervisor) {
	ra.pipeSpec, ra.spec, ra.super = pipeSpec, pipeSpec.FilterSpec().(*Spec), super
	ra.reload()
}

// Inherit inherits previous generation of RequestAdaptor.
func (ra *RequestAdaptor) Inherit(pipeSpec *httppipeline.FilterSpec,
	previousGeneration httppipeline.Filter, super *supervisor.Supervisor) {

	previousGeneration.Close()
	ra.Init(pipeSpec, super)
}

func (ra *RequestAdaptor) reload() {
	if ra.spec.Path != nil {
		ra.pa = pathadaptor.New(ra.spec.Path)
	}
}

// Handle adapts request.
func (ra *RequestAdaptor) Handle(ctx context.HTTPContext) string {
	result := ra.handle(ctx)
	return ctx.CallNextHandler(result)
}

func (ra *RequestAdaptor) handle(ctx context.HTTPContext) string {
	r := ctx.Request()
	method, path, header := r.Method(), r.Path(), r.Header()

	if ra.spec.Method != "" && ra.spec.Method != method {
		ctx.AddTag(stringtool.Cat("requestAdaptor: method ",
			method, " adapted to ", ra.spec.Method))
		r.SetMethod(ra.spec.Method)
	}
	if ra.pa != nil {
		adaptedPath := ra.pa.Adapt(path)
		if adaptedPath != path {
			ctx.AddTag(stringtool.Cat("requestAdaptor: path ",
				path, " adapted to ", adaptedPath))
		}
		r.SetPath(adaptedPath)
	}
	hte := ctx.Template()
	if ra.spec.Header != nil {
		header.Adapt(ra.spec.Header, hte)
	}

	if len(ra.spec.Body) != 0 {
		if hte.HasTemplates(ra.spec.Body) {
			if body, err := hte.Render(ra.spec.Body); err != nil {
				logger.Errorf("BUG resquest render body faile , template %s , err %v",
					ra.spec.Body, err)
			} else {
				ctx.Request().SetBody(bytes.NewReader([]byte(body)))
			}
		} else {
			ctx.Request().SetBody(bytes.NewReader([]byte(ra.spec.Body)))
		}
	}

	return ""
}

// Status returns status.
func (ra *RequestAdaptor) Status() interface{} { return nil }

// Close closes RequestAdaptor.
func (ra *RequestAdaptor) Close() {}