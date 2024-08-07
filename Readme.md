# Demo Stripe Payment system

This is a simple test with HTML, JavaScript and Go

This demo use Stripe as payment gateway

```
Index(HTML) -> Create a Payment Intent(JS) -> Client Secret(GO)
Client Secret(HTML) -> Confirm Payment(JS) -> Payment Intent(GO)
Success(JS) -> Redirect to success page(HTML)
```

## To Run Local

1. Install Go
2. Install Stripe CLI
3. Run `stripe listen --forward-to localhost:9002/webhook`
4. Run `go run main.go`
5. Open `http://localhost:9002`
6. Use card numbers to test
7. Use any future date as expiration date
8. Use any 3 digits as CVC


## Card Numbers to Test
https://docs.stripe.com/payments/accept-a-payment?platform=web&ui=elements#web-test-the-integration

| Card Number            | Scenario                                                    | How to Test                                                                                           |
|------------------------|-------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| 4242424242424242       | The card payment succeeds and doesn’t require authentication. | Fill out the credit card form using the credit card number with any expiration, CVC, and postal code. |
| 4000002500003155       | The card payment requires authentication.                   | Fill out the credit card form using the credit card number with any expiration, CVC, and postal code. |
| 4000000000009995       | The card is declined with a decline code like insufficient_funds. | Fill out the credit card form using the credit card number with any expiration, CVC, and postal code. |
| 6205500000000000004    | The UnionPay card has a variable length of 13-19 digits.    | Fill out the credit card form using the credit card number with any expiration, CVC, and postal code. |


## Links and Docs:

- https://stripe.com/docs/payments/accept-a-payment
- https://stripe.com/docs/payments/accept-a-payment?platform=web&ui=elements#web-test-the-integration
- https://docs.stripe.com/payments/accept-a-payment?ui=elements