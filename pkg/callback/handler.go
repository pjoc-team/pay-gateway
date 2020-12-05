package callback

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pjoc-team/tracing/logger"
	"net/http"
)

// StartGin start gin server
func StartGin(ctx context.Context, service *NotifyService, listenAddr string) {
	engine := gin.New()
	engine.GET("/notify/:gateway_order_id", handleGatewayOrderIDFunc(service)).
		POST("/notify/:gateway_order_id", handleGatewayOrderIDFunc(service))
	err := engine.Run(listenAddr)
	if err != nil{
		log := logger.ContextLog(ctx)
		log.Fatal(err.Error())
	}
}

func handleGatewayOrderIDFunc(service *NotifyService) func(*gin.Context) {
	return func(context *gin.Context) {
		log := logger.ContextLog(context)

		gatewayOrderID := context.Param("gateway_order_id")
		if gatewayOrderID == "" {
			log.Errorf("No parameter gateway_order_id found! request: %v", context.Params)
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		log.Infof("Processing notify: %s", gatewayOrderID)
		request := context.Request
		notifyResponse, e := service.Notify(context, gatewayOrderID, request)
		if e != nil {
			log.Errorf("Failed to process notify! orderId: %s error: %s", gatewayOrderID, e.Error())
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		response := notifyResponse.Response
		headers := response.Header
		context.Status(int(response.Status))
		for name, value := range headers {
			context.Header(name, value)
		}
		if n, e := context.Writer.Write(response.Body); e != nil {
			log.Errorf("failed to write response! error: %v", e.Error())
		} else {
			log.Debugf("Success response with size: %d", n)
		}
	}
}
