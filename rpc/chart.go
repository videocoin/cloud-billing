package rpc

import (
	"time"

	v1 "github.com/videocoin/cloud-api/billing/v1"
)

func NewFillChartChargesResponse() *v1.ChartChargesResponse {
	resp := &v1.ChartChargesResponse{Items: []*v1.ChartChargeResponse{}}

	now := time.Now()
	countOfDays := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	curDay := 1
	for {
		if curDay > countOfDays {
			break
		}

		item := &v1.ChartChargeResponse{
			Name: time.Date(now.Year(), now.Month(), curDay, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Live: 0,
			Vod:  0,
		}

		resp.Items = append(resp.Items, item)

		curDay++
	}

	return resp
}

func FillChartChargesResponseWithData(resp *v1.ChartChargesResponse, charges []*v1.ChargeResponse) {
	for _, item := range resp.Items {
		d, _ := time.Parse(time.RFC3339, item.Name)
		item.Live = CalcChargeAmountByDate(charges, true, d)
		item.Vod = CalcChargeAmountByDate(charges, false, d)
	}
}

func CalcChargeAmountByDate(charges []*v1.ChargeResponse, isLive bool, d time.Time) float64 {
	value := float64(0)

	for _, item := range charges {
		if isLive && !item.StreamIsLive {
			continue
		}

		if !isLive && item.StreamIsLive {
			continue
		}

		if d.Year() == item.CreatedAt.Year() && d.Month() == item.CreatedAt.Month() && d.Day() == item.CreatedAt.Day() {
			value += item.TotalCost
		}
	}

	return value
}
