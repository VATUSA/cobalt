package background

import (
	"context"
	"fmt"
	"io/ioutil"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Job struct {
	Name string
	Args []string
}

func NewJob(name string, args ...string) *Job {
	return &Job{
		Name: name,
		Args: args,
	}
}

func (j *Job) SetArgs(args ...string) {
	j.Args = args
}

func (j *Job) Run() error {
	fmt.Printf("Running job %s with args %v\n", j.Name, j.Args)

	err := runBackgroundJob(j.Name, j.Args)
	if err != nil {
		return err
	}
	return nil
}

func getK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func getCurrentNamespace() string {
	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "cobalt"
	}
	return string(data)
}

func runBackgroundJob(jobName string, args []string) error {
	cmdArgs := []string{
		"/app/cobalt-background",
		jobName,
	}
	cmdArgs = append(cmdArgs, args...)

	name := fmt.Sprintf("cobalt-bg-%s", jobName)

	err := launchK8sJob(name, cmdArgs)
	if err != nil {
		return err
	}
	return nil
}

func launchK8sJob(name string, cmdArgs []string) error {
	clientSet, err := getK8sClient()
	if err != nil {
		return err
	}
	namespace := getCurrentNamespace()
	deployments := clientSet.AppsV1().Deployments(namespace)
	cobaltDeployment, err := deployments.Get(context.Background(), "cobalt", metav1.GetOptions{})
	if err != nil {
		return err
	}
	image := cobaltDeployment.Spec.Template.Spec.Containers[0].Image
	jobs := clientSet.BatchV1().Jobs(namespace)

	var backoffLimit int32 = 0
	var ttlSecondsAfterFinished int32 = 86400 // Delete job 1 day after completion

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    name,
							Image:   image,
							Command: cmdArgs,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}

	_, err = jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}
