const stripe = Stripe("pk_xxxyyyzzz");

let elements;

document.querySelector("#payment-form").addEventListener("submit", handleSubmit);

async function validateEmail() {
    const email = document.querySelector("#email").value;
    const course_name = document.querySelector("#course_name").value;
    const course_id = document.querySelector("#course_id").value;
    const price = document.querySelector("#price-unformatted").value;
    if (validateEmailFormat(email)) {
        setLoading(true);
        const response = await fetch("http://localhost:9002/create-payment-intent", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email, course_name, course_id, price }),
        });
        const { clientSecret } = await response.json();
        const appearance = { theme: 'stripe' };
        elements = stripe.elements({ appearance, clientSecret });
        const paymentElementOptions = { layout: "tabs" };
        const paymentElement = elements.create("payment", paymentElementOptions);
        paymentElement.mount("#payment-element");
        document.querySelector("#button-text").textContent = "Pay Now";
        document.querySelector("#submit").disabled = false;
        setLoading(false);
    } else {
        document.querySelector("#submit").disabled = true;
    }
}

function validateEmailFormat(email) {
    const re = /\S+@\S+\.\S+/;
    return re.test(email);
}

async function handleSubmit(e) {
    e.preventDefault();

    const formattedPrice = document.getElementById("price").value;
    const unformattedPrice = formattedPrice.replace(/[^0-9,]/g, '').replace(',', '.');
    document.getElementById("price-unformatted").value = unformattedPrice;

    if (document.querySelector("#button-text").textContent === "Check") {
        await validateEmail();
    } else {
        setLoading(true);
        const { error } = await stripe.confirmPayment({
            elements,
            confirmParams: {
                return_url: "http://localhost:9002/index.html",
            },
        });

        if (error.type === "card_error" || error.type === "validation_error") {
            showMessage(error.message, false);
        } else {
            showMessage("An unexpected error occurred.", false);
        }
        setLoading(false);
    }
}

async function checkStatus() {
    const clientSecret = new URLSearchParams(window.location.search).get("payment_intent_client_secret");
    if (!clientSecret) {
        return;
    }

    const { paymentIntent } = await stripe.retrievePaymentIntent(clientSecret);
    switch (paymentIntent.status) {
        case "succeeded":

            // hide all form, show a banner with success and redirect
            document.querySelector("#payment-form").classList.add("hidden");
            document.querySelector("#payment-message-success").classList.remove("hidden");

            // set time out 4s
            setTimeout(() => {
                window.location.href = "http://localhost:9002";

            }, 4000);
            break;
        case "processing":
            showMessage("Your payment is processing.", true);
            break;
        case "requires_payment_method":
            showMessage("Your payment was not successful, please try again.", false);
            break;
        default:
            showMessage("Something went wrong.", false);
            break;
    }
}

function showMessage(messageText, isSuccess = true) {
    const messageContainer = document.querySelector("#payment-message");
    messageContainer.classList.remove("hidden", "alert-success", "alert-danger");3

    if (isSuccess) {
        messageContainer.classList.add("alert", "alert-success");
    } else {
        messageContainer.classList.add("alert", "alert-danger");
    }

    messageContainer.textContent = messageText;

    setTimeout(() => {
        messageContainer.classList.add("hidden");
        messageContainer.textContent = "";
        messageContainer.classList.remove("alert", "alert-success", "alert-danger");
    }, 5000);
}

function setLoading(isLoading) {
    if (isLoading) {
        document.querySelector("#submit").disabled = true;
        document.querySelector("#spinner").classList.remove("hidden");
        document.querySelector("#button-text").classList.add("hidden");
    } else {
        document.querySelector("#submit").disabled = false;
        document.querySelector("#spinner").classList.add("hidden");
        document.querySelector("#button-text").classList.remove("hidden");
    }
}

checkStatus();
