package dbservice

import (
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"github.com/pjoc-team/pay-gateway/pkg/constant"
	"github.com/pjoc-team/pay-gateway/pkg/dbservice/model"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"golang.org/x/net/context"
)

// PayDatabaseService service of db
type PayDatabaseService struct {
	*gorm.DB
}

// FindPayNotifyLessThenTime find notifys less then time
func (s *PayDatabaseService) FindPayNotifyLessThenTime(
	ctx context.Context, payNotify *pb.PayNotify,
) (response *pb.PayNotifyResponse, err error) {
	log := logger.ContextLog(ctx)
	notify := newDbNotify(payNotify)
	// if err = copier.Copy(notify, payNotify); err != nil {
	//	log.Errorf("failed to copy object! error: %s", err)
	//	return
	// }
	results := make([]model.Notify, 0)
	if results := s.Where(
		"length(next_notify_time) > 0 and next_notify_time <= ? and status != ?",
		notify.NextNotifyTime, constant.NotifySuccess,
	).Find(&results); results.RecordNotFound() {
		log.Errorf("find error: %v", s.Error.Error())
		return
	}
	response = &pb.PayNotifyResponse{}
	response.PayNotifys = make([]*pb.PayNotify, len(results))
	for i, notify := range results {
		payNotify := &pb.PayNotify{}
		if err = copier.Copy(payNotify, notify); err != nil {
			log.Error("copy result error! error: %v", err.Error())
		} else {
			log.Debugf("found result: %v by query: %v", response, payNotify)
		}
		response.PayNotifys[i] = payNotify
	}
	return
}

// SavePayNotify save pay notify data
func (s *PayDatabaseService) SavePayNotify(
	ctx context.Context, payNotify *pb.PayNotify,
) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	notify := newDbNotify(payNotify)
	if dbResult := s.Create(notify); dbResult.Error != nil {
		log.Errorf("failed to save notify! notify: %v error: %s", payNotify, err.Error())
		err = dbResult.Error
		return
	}
	log.Infof("succeed save notify: %v", payNotify)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

// UpdatePayNotify update notify by id
func (s *PayDatabaseService) UpdatePayNotify(
	ctx context.Context, payNotify *pb.PayNotify,
) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	notify := newDbNotify(payNotify)
	if dbResult := s.Model(notify).Update(notify); dbResult.Error != nil {
		err = dbResult.Error
		log.Errorf("failed to update notify! notify: %v error: %s", payNotify, err.Error())
		return
	}
	log.Infof("succeed update notify: %v", payNotify)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

// FindPayNotify find notify
func (s *PayDatabaseService) FindPayNotify(
	ctx context.Context, payNotify *pb.PayNotify,
) (response *pb.PayNotifyResponse, err error) {
	log := logger.ContextLog(ctx)
	notify := newDbNotify(payNotify)
	results := make([]model.Notify, 0)
	if results := s.Find(&results, notify); results.RecordNotFound() {
		log.Errorf("find error: %v", s.Error.Error())
		return
	}
	response = &pb.PayNotifyResponse{}
	response.PayNotifys = make([]*pb.PayNotify, len(results))
	for i, notify := range results {
		payNotify := &pb.PayNotify{}
		if err = copier.Copy(payNotify, notify); err != nil {
			log.Error("copy result error! error: %v", err.Error())
		} else {
			log.Debugf("found result: %v by query: %v", response, payNotify)
		}
		response.PayNotifys[i] = payNotify
	}

	return
}

// SavePayNotifyOk save notify ok data
func (s *PayDatabaseService) SavePayNotifyOk(
	ctx context.Context, payNotifyOkRequest *pb.PayNotifyOk,
) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	tx := s.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	notifyOk := newDbNotifyOk(payNotifyOkRequest)
	if dbResult := s.Create(notifyOk); dbResult.Error != nil {
		log.Errorf(
			"failed to save ok order! order: %v error: %s", payNotifyOkRequest,
			dbResult.Error.Error(),
		)
		err = dbResult.Error
		tx.Rollback()
		return
	}
	notify := &model.Notify{GatewayOrderID: payNotifyOkRequest.GatewayOrderId}
	notify.Status = constant.OrderStatusSuccess
	if update := s.Model(notify).Update(notify); update.Error != nil {
		log.Errorf("failed to update notify!")
		tx.Rollback()
		return
	}
	err = tx.Commit().Error

	log.Infof("succeed save ok notify: %v", payNotifyOkRequest)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

// FindPayNotifyOk find notify ok data
func (s *PayDatabaseService) FindPayNotifyOk(
	ctx context.Context, payNotifyOk *pb.PayNotifyOk,
) (response *pb.PayNotifyOkResponse, err error) {
	log := logger.ContextLog(ctx)
	notifyOk := newDbNotifyOk(payNotifyOk)
	results := make([]model.NotifyOk, 0)
	if results := s.Find(&results, notifyOk); results.RecordNotFound() {
		log.Errorf("find error: %v", s.Error.Error())
		return
	}
	response = &pb.PayNotifyOkResponse{}
	response.PayNotifyOks = make([]*pb.PayNotifyOk, len(results))

	for i, notifyOk := range results {
		payNotifyOk := &pb.PayNotifyOk{}
		if err = copier.Copy(payNotifyOk, notifyOk); err != nil {
			log.Error("copy result error! error: %v", err.Error())
		} else {
			log.Debugf("found result: %v by query: %v", response, payNotifyOk)
		}
		response.PayNotifyOks[i] = payNotifyOk
	}

	if err = copier.Copy(&response.PayNotifyOks, results); err != nil {
		log.Error("copy result error! error: %v", err.Error())
	} else {
		log.Debugf("found result: %v by query: %v", response, payNotifyOk)
	}
	return
}

// UpdatePayNotifyOk update pay notify ok data
func (s *PayDatabaseService) UpdatePayNotifyOk(
	ctx context.Context, payNotifyOk *pb.PayNotifyOk,
) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	notifyOk := newDbNotifyOk(payNotifyOk)
	if dbResult := s.Model(notifyOk).Update(notifyOk); dbResult.Error != nil {
		log.Errorf("failed to save ok notify! notifyOk: %v error: %s", payNotifyOk, err.Error())
		err = dbResult.Error
		return
	}
	log.Infof("succeed save ok notify: %v", payNotifyOk)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

// FindPayOrder find pay order
func (s *PayDatabaseService) FindPayOrder(
	ctx context.Context, orderRequest *pb.PayOrder,
) (response *pb.PayOrderResponse, err error) {
	log := logger.ContextLog(ctx)
	order := newDbPayOrder(orderRequest)
	results := make([]model.PayOrder, 0)
	if results := s.Find(&results, order); results.RecordNotFound() {
		log.Errorf("find error: %v", s.Error.Error())
		return
	}
	if log.IsDebugEnabled() {
		log.Debugf("find order: %v by order: %v", results, orderRequest)
	}
	response = &pb.PayOrderResponse{}
	response.PayOrders = make([]*pb.PayOrder, len(results))
	for i, payOrder := range results {
		order := &pb.PayOrder{}
		order.BasePayOrder = &pb.BasePayOrder{}
		response.PayOrders[i] = order

		if err = copier.Copy(response.PayOrders[i], payOrder); err != nil {
			log.Error("copy result error! error: %v", err.Error())
		} else if err = copier.Copy(order.BasePayOrder, payOrder); err != nil {
			log.Error("copy result error! error: %v", err.Error())
		} else {
			log.Debugf("found result: %v by query: %v", response, orderRequest)
		}
	}

	return
}

// FindPayOrderOk find pay order ok data
func (s *PayDatabaseService) FindPayOrderOk(
	ctx context.Context, orderOkRequest *pb.PayOrderOk,
) (response *pb.PayOrderOkResponse, err error) {
	log := logger.ContextLog(ctx)
	orderOk := newDbPayOrderOk(orderOkRequest)
	results := make([]model.PayOrderOk, 0)
	if results := s.Find(&results, orderOk); results.RecordNotFound() {
		log.Errorf("find error: %v", s.Error.Error())
		return
	}
	response = &pb.PayOrderOkResponse{}
	response.PayOrderOks = make([]*pb.PayOrderOk, len(results))
	for i, payOrderOk := range results {
		orderOk := &pb.PayOrderOk{}
		orderOk.BasePayOrder = &pb.BasePayOrder{}
		response.PayOrderOks[i] = orderOk

		if err = copier.Copy(orderOk, payOrderOk); err != nil {
			log.Error("copy result error! error: %v", err.Error())
		} else if err = copier.Copy(orderOk.BasePayOrder, payOrderOk); err != nil {
			log.Error("copy result error! error: %v", err.Error())
		} else {
			log.Debugf("found result: %v by query: %v", response, orderOkRequest)
		}
	}
	return
}

// SavePayOrder save pay order
func (s *PayDatabaseService) SavePayOrder(
	ctx context.Context, orderRequest *pb.PayOrder,
) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	order := newDbPayOrder(orderRequest)
	if dbResult := s.Create(order); dbResult.Error != nil {
		log.Errorf(
			"failed to save order! order: %v error: %s", orderRequest, dbResult.Error.Error(),
		)
		err = dbResult.Error
		return
	}
	log.Infof("succeed save order: %v", orderRequest)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

// UpdatePayOrder update pay order
func (s *PayDatabaseService) UpdatePayOrder(
	ctx context.Context, orderRequest *pb.PayOrder,
) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	order := newDbPayOrder(orderRequest)
	if dbResult := s.Model(order).Update(order); dbResult.Error != nil {
		log.Errorf(
			"failed to update order! order: %v error: %s", orderRequest, dbResult.Error.Error(),
		)
		err = dbResult.Error
		return
	}
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	log.Infof("succeed update order: %v result: %v", orderRequest, result)
	return
}

// SavePayOrderOk save pay order ok data
func (s *PayDatabaseService) SavePayOrderOk(
	ctx context.Context, orderOkRequest *pb.PayOrderOk,
) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	tx := s.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	order := newDbPayOrderOk(orderOkRequest)
	if dbResult := s.Create(order); dbResult.Error != nil {
		log.Errorf(
			"failed to save ok order! order: %v error: %s", orderOkRequest, dbResult.Error.Error(),
		)
		err = dbResult.Error
		tx.Rollback()
		return
	}
	payOrder := &model.PayOrder{BasePayOrder: model.BasePayOrder{GatewayOrderID: orderOkRequest.BasePayOrder.GatewayOrderId}}
	payOrder.OrderStatus = constant.OrderStatusSuccess
	if update := s.Model(payOrder).Update(payOrder); update.Error != nil {
		log.Errorf("failed to update order!")
		tx.Rollback()
		return
	}
	err = tx.Commit().Error

	log.Infof("succeed save ok order: %v", orderOkRequest)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

// UpdatePayOrderOk update pay order ok data
func (s *PayDatabaseService) UpdatePayOrderOk(
	ctx context.Context, orderOkRequest *pb.PayOrderOk,
) (result *pb.ReturnResult, err error) {
	log := logger.ContextLog(ctx)
	order := newDbPayOrderOk(orderOkRequest)
	if dbResult := s.Model(order).Update(order); dbResult.Error != nil {
		log.Errorf(
			"failed to save ok order! order: %v error: %s", orderOkRequest, dbResult.Error.Error(),
		)
		err = dbResult.Error
		return
	}
	log.Infof("succeed save ok order: %v", orderOkRequest)
	result = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS}
	return
}

// NewServer new database service
func NewServer(db *gorm.DB) (pb.PayDatabaseServiceServer, error) {
	svc := &PayDatabaseService{
		DB: db,
	}
	return svc, nil
}
