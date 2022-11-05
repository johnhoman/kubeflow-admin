package pod

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/crossplane/crossplane-runtime/pkg/event"
    "github.com/crossplane/crossplane-runtime/pkg/logging"
    corev1 "k8s.io/api/core/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type HandlerOption func(h *handler)

func WithLogger(logger logging.Logger) HandlerOption {
    return func(h *handler) {
        h.logger = logger
    }
}

func WithEventRecorder(ev event.Recorder) HandlerOption {
    return func(h *handler) {
        h.record = ev
    }
}

func WithMutateFunc(fn MutateFunc) HandlerOption {
    return func(h *handler) {
        h.mutateFunc = fn
    }
}

func WithReader(cli client.Reader) HandlerOption {
    return func(h *handler) {
        h.reader = cli
    }
}

func WithPredicate(fn PredicateFunc) HandlerOption {
    return func(h *handler) {
        h.predicateFunc = fn
    }
}

type MutateFunc func(ctx context.Context, reader client.Reader, pod *corev1.Pod) error
type PredicateFunc func(pod *corev1.Pod) bool

func NewHandler(opts ...HandlerOption) *handler {
    h := &handler{
        reader: nil,
        logger: logging.NewNopLogger(),
        record: event.NewNopRecorder(),
    }
    for _, f := range opts {
        f(h)
    }
    return h
}

type handler struct {
    reader client.Reader
    logger logging.Logger
    record event.Recorder
    decoder *admission.Decoder
    mutateFunc MutateFunc
    predicateFunc PredicateFunc
}

func (h *handler) Handle(ctx context.Context, req admission.Request) admission.Response {

    pod := &corev1.Pod{}
    if err := h.decoder.Decode(req, pod); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }

    if h.predicateFunc != nil && !h.predicateFunc(pod) {
        return admission.Allowed("ignored")
    }

    if err := h.mutateFunc(ctx, h.reader, pod); err != nil {
        return admission.Errored(http.StatusInternalServerError, err)
    }

    raw, err := json.Marshal(pod)
    if err != nil {
        return admission.Errored(http.StatusInternalServerError, err)
    }

    return admission.PatchResponseFromRaw(req.Object.Raw, raw)
}


func (h *handler) InjectDecoder(decoder *admission.Decoder) error {
    h.decoder = decoder
    return nil
}

var _ admission.Handler = &handler{}
var _ admission.DecoderInjector = &handler{}
