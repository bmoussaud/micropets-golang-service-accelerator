SOURCE_IMAGE = os.getenv("SOURCE_IMAGE", default='imageRegistry/micropet-tap-lowercasePetKind-sources')
LOCAL_PATH = os.getenv("LOCAL_PATH", default='.')
NAMESPACE = os.getenv("NAMESPACE", default='dev-tap')

compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/cmd -buildmode pie -trimpath ./cmd/main.go'

allow_k8s_contexts('aks-eu-tap-7')

local_resource(
  'go-build',
  compile_cmd,
  deps=['./cmd', './service','./internal'],
  dir='.')

k8s_custom_deploy(
    'lowercasePetKind',
    apply_cmd="tanzu apps workload apply -f config/workload.yaml --live-update" +
               " --local-path " + LOCAL_PATH +
               " --source-image " + SOURCE_IMAGE +
               " --namespace " + NAMESPACE +
               " --yes >/dev/null" +
               " && kubectl get workload lowercasePetKind --namespace " + NAMESPACE + " -o yaml",
    delete_cmd="tanzu apps workload delete -f config/workload.yaml --namespace " + NAMESPACE + " --yes",    
    deps=['./build'],
    container_selector='workload',
    live_update=[      
      sync('./build', '/tmp/tilt')  ,      
      run('cp -rf /tmp/tilt/* /layers/tanzu-buildpacks_go-build/targets/bin', trigger=['./build']),
    ]
)

k8s_resource('lowercasePetKind', port_forwards=["8080:8080"], extra_pod_selectors=[{'serving.knative.dev/service': 'lowercasePetKind-java'}]) 

