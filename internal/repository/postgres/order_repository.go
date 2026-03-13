package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

func (r *OrderRepository) GetByUID(ctx context.Context, orderUID string) (domain.Order, error) {
	var or orderRow
	err := r.pool.QueryRow(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature,
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1
	`, orderUID).Scan(
		&or.OrderUID,
		&or.TrackNumber,
		&or.Entry,
		&or.Locale,
		&or.InternalSignature,
		&or.CustomerID,
		&or.DeliveryService,
		&or.ShardKey,
		&or.SmID,
		&or.DateCreated,
		&or.OofShard,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, ErrOrderNotFound
		}
		return domain.Order{}, err
	}

	var dr deliveryRow
	err = r.pool.QueryRow(ctx, `
		SELECT order_uid, name, phone, zip, city, address, region, email
		FROM deliveries
		WHERE order_uid = $1
	`, orderUID).Scan(
		&dr.OrderUID,
		&dr.Name,
		&dr.Phone,
		&dr.Zip,
		&dr.City,
		&dr.Address,
		&dr.Region,
		&dr.Email,
	)
	if err != nil {
		return domain.Order{}, err
	}

	var pr paymentRow
	err = r.pool.QueryRow(ctx, `
		SELECT transaction, order_uid, request_id, currency, provider, amount,
		       payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments
		WHERE order_uid = $1
	`, orderUID).Scan(
		&pr.Transaction,
		&pr.OrderUID,
		&pr.RequestID,
		&pr.Currency,
		&pr.Provider,
		&pr.Amount,
		&pr.PaymentDT,
		&pr.Bank,
		&pr.DeliveryCost,
		&pr.GoodsTotal,
		&pr.CustomFee,
	)
	if err != nil {
		return domain.Order{}, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, order_uid, chrt_id, track_number, price, rid, name,
		       sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1
		ORDER BY id
	`, orderUID)
	if err != nil {
		return domain.Order{}, err
	}
	defer rows.Close()

	items := make([]itemRow, 0)
	for rows.Next() {
		var ir itemRow
		if err := rows.Scan(
			&ir.ID,
			&ir.OrderUID,
			&ir.ChrtID,
			&ir.TrackNumber,
			&ir.Price,
			&ir.RID,
			&ir.Name,
			&ir.Sale,
			&ir.Size,
			&ir.TotalPrice,
			&ir.NmID,
			&ir.Brand,
			&ir.Status,
		); err != nil {
			return domain.Order{}, err
		}
		items = append(items, ir)
	}

	if err := rows.Err(); err != nil {
		return domain.Order{}, err
	}

	return buildOrder(or, dr, pr, items), nil
}

func (r *OrderRepository) ListRecent(ctx context.Context, limit, offset int) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT order_uid
		FROM orders
		ORDER BY date_created DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderUIDs := make([]string, 0, limit)
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, err
		}
		orderUIDs = append(orderUIDs, orderUID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	orders := make([]domain.Order, 0, len(orderUIDs))
	for _, orderUID := range orderUIDs {
		order, err := r.GetByUID(ctx, orderUID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}