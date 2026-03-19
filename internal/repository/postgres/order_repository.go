package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrOrderDuplicated = errors.New("order duplicated")
)

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

func (r *OrderRepository) Save(ctx context.Context, order domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (
			order_uid,
			track_number,
			entry,
			locale,
			internal_signature,
			customer_id,
			delivery_service,
			shardkey,
			sm_id,
			date_created,
			oof_shard
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11
		)
	`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.ShardKey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrOrderDuplicated
		}
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO deliveries (
			order_uid,
			name,
			phone,
			zip,
			city,
			address,
			region,
			email
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (
			transaction,
			order_uid,
			request_id,
			currency,
			provider,
			amount,
			payment_dt,
			bank,
			delivery_cost,
			goods_total,
			custom_fee
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11
		)
	`,
		order.Payment.Transaction,
		order.OrderUID,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDT,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO items (
				order_uid,
				chrt_id,
				track_number,
				price,
				rid,
				name,
				sale,
				size,
				total_price,
				nm_id,
				brand,
				status
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11, $12
			)
		`,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.RID,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}