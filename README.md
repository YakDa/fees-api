# Fees API

A billing fees API built with **Encore** and **Temporal** for the Pave Bank coding challenge.

## Tech Stack

- **Encore** - Backend framework for building APIs
- **Temporal** - Workflow orchestration for billing periods
- **Go** - Programming language

## Architecture

This project follows a 3-layer clean architecture:

```
fees-api/
├── cmd/api/                    # Entry point (Encore)
├── internal/
│   ├── model/                  # Data models
│   ├── presentation/handlers/  # API handlers
│   ├── service/                # Business logic
│   └── repository/             # Data access
├── pkg/errors/                 # Error types
├── billing/                    # Encore API service
├── workflow/                   # Temporal workflow
├── encore.app
├── go.mod
└── README.md
```

- **Model** - Data structures (Bill, LineItem, DTOs)
- **Presentation** - HTTP handlers (Encore API endpoints)
- **Service** - Business logic (currency conversion, validation)
- **Repository** - Data storage (in-memory)

## API Endpoints

### Create Bill
```bash
POST /billing.CreateBill
{
  "currency": "USD"  # or "GEL"
}
```

### Add Line Item
```bash
POST /billing.AddLineItem?billId={bill_id}
{
  "description": "Service fee",
  "amount": 10.00,
  "currency": "USD"  # or "GEL"
}
```

### Close Bill
```bash
POST /billing.CloseBill?billId={bill_id}
```

### Get Bill
```bash
GET /billing.GetBill?billId={bill_id}
```

### List Bills
```bash
GET /billing.ListBills?status=open
GET /billing.ListBills?status=closed
```

## Features

- Create new bills
- Add line items to open bills
- Close active bills (shows total + line items)
- Reject line items on closed bills
- Query open and closed bills
- Multi-currency support (GEL, USD)
- Currency conversion for totals
- Temporal workflow for billing periods
- Event-driven architecture

## Running Locally

```bash
# Install Encore CLI
encore run
```

## Design Decisions

### Money Representation
- Amounts stored as float64 for simplicity
- Currency explicitly tracked per bill and line item
- Conversion to USD for display (exchange rates configurable)

### Data Model
- Bill - Contains status, currency, total amount, line items
- LineItem - Description, amount, currency, timestamps

### Temporal Workflow
- Workflow started when billing period begins
- Progressive accrual of fees via signals
- Automatic billing period end via timer
- Queryable state for monitoring

### Why Temporal?
- Reliability - Ensures billing operations complete even on failures
- Auditability - Complete history of bill changes
- Temporal consistency - No race conditions on bill modifications

## Notes

- In production, replace in-memory storage with a database (PostgreSQL via Encore)
- Exchange rates should come from a real-time service
- Add authentication and authorization
- Implement proper error handling and logging
