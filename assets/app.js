const APP_ID = '';
const APP_KEY = '';

const submitButton = document.querySelector('#submit');
const status = {
  '0': '欄位已填妥',
  '1': '欄位未填妥',
  '2': '欄位有錯誤',
  '3': '輸入中',
};
const style = {
  color: 'black',
  fontSize: '16px',
  lineHeight: '24px',
  fontWeight: '300',
  errorColor: 'red',
  placeholderColor: '',
};
const config = {
  isUsedCcv: true,
};

TPDirect.setupSDK(APP_ID, APP_KEY, 'sandbox');
TPDirect.card.setup('#tappay-iframe', style, config);
TPDirect.card.onUpdate((update) => {
  document.getElementById('message').innerHTML = `
    Card Number Status: ${status[update.status.number]} <br>
    Card Expiry Status: ${status[update.status.expiry]} <br>
    Cvc Status: ${status[update.status.ccv]}
  `;
  update.canGetPrime
    ? submitButton.removeAttribute('disabled')
    : submitButton.setAttribute('disabled', 'true');
});

submitButton.addEventListener('click', () => {
  TPDirect.card.getPrime(async (result) => {
    const res = await pay({
      prime: result.card.prime,
    });
    document.getElementById('message').innerHTML = `
      Message: ${res.msg} <br>
      Amount: ${res.amount} <br>
      Currency: ${res.currency} <br>
      Merchant ID: ${res.merchant_id}
    `;
    console.log(res);
    submitButton.setAttribute('hidden', 'true');
  });
});

const pay = async (data) => {
  try {
    const res = await fetch('/api/pay', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return res.json();
  } catch {
    //
  }
};
