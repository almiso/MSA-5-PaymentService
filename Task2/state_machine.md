# Таблица переходов состояний платежа (State Machine)

| Исходное состояние             | Переходное состояние     | Событие                                                                |
|:-------------------------------|:-------------------------|:-----------------------------------------------------------------------|
| `NONE` (Или отсутствие записи) | `PENDING`                | `PAYMENT_CREATED` (Пользователь нажал "Оплатить")                      |
| `PENDING`                      | `MONEY_HELD`             | `MONEY_CHARGED_SUCCESS` (Средства успешно захолдированы/списаны)       |
| `MONEY_HELD`                   | `FRAUD_APPROVED`         | `FRAUD_CHECK_PASSED` (Антифрод: "Разрешение")                          |
| `MONEY_HELD`                   | `FRAUD_REJECTED`         | `FRAUD_CHECK_FAILED` (Антифрод: "Запрет")                              |
| `MONEY_HELD`                   | `MANUAL_REVIEW`          | `FRAUD_CHECK_MANUAL` (Антифрод: "Ручная проверка")                     |
| `MANUAL_REVIEW`                | `FRAUD_APPROVED`         | `OPERATOR_APPROVED` (Оператор разрешил операцию)                       |
| `MANUAL_REVIEW`                | `FRAUD_APPROVED`         | `CUT_OFF_TIMEOUT` (Сработал таймер cut-off)                            |
| `MANUAL_REVIEW`                | `FRAUD_REJECTED`         | `OPERATOR_REJECTED` (Оператор запретил операцию)                       |
| `FRAUD_APPROVED`               | `COMPLETED`              | `TRANSFER_SUCCESS` (Деньги успешно ушли контрагенту, финальный статус) |
| `FRAUD_REJECTED`               | `REFUND_PENDING`         | `COMPENSATION_STARTED` (Запуск Saga-компенсации из-за неудачи/запрета) |
| `REFUND_PENDING`               | `CANCELED`               | `REFUND_SUCCESS` (Деньги успешно возвращены, финальный статус)         |
