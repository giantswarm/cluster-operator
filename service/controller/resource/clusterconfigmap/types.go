package clusterconfigmap

type configMapSpec struct {
	Name        string
	Namespace   string
	Values      map[string]interface{}
	Labels      map[string]string
	Annotations map[string]string
}
