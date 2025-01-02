package imager_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/Muchogoc/imager/pkg/imager"
)

func TestImager_GetChartImagesDetails(t *testing.T) {
	type args struct {
		ctx      context.Context
		chartURL string
	}

	tests := []struct {
		name    string
		args    args
		want    []*imager.ImageDetails
		wantErr bool
	}{
		{
			name: "happy case: http scheme helm chart",
			args: args{
				ctx:      context.Background(),
				chartURL: "https://github.com/openfga/helm-charts/releases/download/openfga-0.2.19/openfga-0.2.19.tgz",
			},
			want: []*imager.ImageDetails{
				{
					Name:   "openfga/openfga:v1.8.2",
					Layers: 4,
					Size:   16.662891387939453,
				},
			},
			wantErr: false,
		},
		{
			name: "happy case: oci scheme helm chart",
			args: args{
				ctx:      context.Background(),
				chartURL: "oci://registry-1.docker.io/bitnamicharts/airflow",
			},
			want: []*imager.ImageDetails{
				{
					Name:   "docker.io/bitnami/airflow:2.10.4-debian-12-r0",
					Layers: 1,
					Size:   448.5532636642456,
				},
			},
			wantErr: false,
		},
		{
			name: "sad case: invalid helm chart link",
			args: args{
				ctx:      context.Background(),
				chartURL: "https://github.com/bonoko",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := imager.NewImager()

			got, err := i.GetChartImagesDetails(tt.args.ctx, tt.args.chartURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Imager.GetChartImagesDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Imager.GetChartImagesDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
