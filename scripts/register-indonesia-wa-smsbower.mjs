#!/usr/bin/env node

const DEFAULTS = {
  smsBaseURL: 'http://byte-v-forge-sms-service:8080/api/sms',
  waBaseURL: 'http://byte-v-forge-wa-app-service:8080/api/wa',
  gopayBaseURL: 'http://byte-v-forge-gopay-app:8080/api/gopay',
  gopayActionBaseURL: '',
  applicationKey: 'wa',
  countryISO2: 'ID',
  countryCallingCode: '62',
  skipGopayCheck: false,
  providerKey: 'smsbower',
  smsbowerCountryID: '6',
  maxAttempts: 10,
  minAvailableCount: 1,
  acquireWaitSeconds: 45,
  otpWaitSeconds: 240,
  attemptDelaySeconds: 0,
};

const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

function parseArgs(argv) {
  const args = {...DEFAULTS};
  for (const raw of argv) {
    const [key, value = ''] = raw.replace(/^--/, '').split('=');
    switch (key) {
      case 'sms-base-url': args.smsBaseURL = value; break;
      case 'wa-base-url': args.waBaseURL = value; break;
      case 'gopay-base-url': args.gopayBaseURL = value; break;
      case 'gopay-action-base-url': args.gopayActionBaseURL = value; break;
      case 'max-attempts': args.maxAttempts = positiveInt(value, args.maxAttempts); break;
      case 'min-available': args.minAvailableCount = positiveInt(value, args.minAvailableCount); break;
      case 'max-price': args.maxPrice = positiveNumber(value); break;
      case 'min-price': args.minPrice = positiveNumber(value); break;
      case 'provider': args.providerKey = value.trim() || args.providerKey; break;
      case 'provider-id': args.providerID = value.trim(); break;
      case 'price': args.price = positiveNumber(value); break;
      case 'country-iso2': args.countryISO2 = value.trim().toUpperCase() || args.countryISO2; break;
      case 'country-calling-code': args.countryCallingCode = String(value || '').replace(/\D/g, '') || args.countryCallingCode; break;
      case 'acquire-wait-seconds': args.acquireWaitSeconds = positiveInt(value, args.acquireWaitSeconds); break;
      case 'otp-wait-seconds': args.otpWaitSeconds = positiveInt(value, args.otpWaitSeconds); break;
      case 'attempt-delay-seconds': args.attemptDelaySeconds = nonNegativeInt(value, args.attemptDelaySeconds); break;
      case 'skip-gopay-check': args.skipGopayCheck = value === '' || value === 'true' || value === '1'; break;
      case 'show-phone': args.showPhone = true; break;
      default: break;
    }
  }
  args.smsBaseURL = env('SMS_API_BASE_URL', args.smsBaseURL);
  args.waBaseURL = env('WA_API_BASE_URL', args.waBaseURL);
  args.gopayBaseURL = env('GOPAY_API_BASE_URL', args.gopayBaseURL);
  args.gopayActionBaseURL = env('GOPAY_ACTION_API_BASE_URL', args.gopayActionBaseURL || `${trimRightSlash(args.gopayBaseURL)}/actions`);
  return args;
}

function trimRightSlash(value) {
  return String(value || '').replace(/\/+$/, '');
}

function env(name, fallback) {
  const value = process.env[name]?.trim();
  return value || fallback;
}

function positiveInt(value, fallback) {
  const parsed = Number.parseInt(String(value || ''), 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function nonNegativeInt(value, fallback) {
  const parsed = Number.parseInt(String(value || ''), 10);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : fallback;
}

function positiveNumber(value) {
  const parsed = Number.parseFloat(String(value || ''));
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : undefined;
}

function requestID(prefix) {
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2, 10)}`;
}

async function jsonFetch(url, options = {}, timeoutMs = 65000) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
      headers: {'Content-Type': 'application/json', ...(options.headers || {})},
    });
    const text = await response.text();
    let body = {};
    try {
      body = text ? JSON.parse(text) : {};
    } catch {
      body = {raw: text};
    }
    return {status: response.status, body};
  } finally {
    clearTimeout(timer);
  }
}

function redactPhone(value, showPhone) {
  const text = String(value || '');
  return showPhone ? text : text.replace(/\d(?=\d{4})/g, '*');
}

function phoneFromOrder(order, config) {
  const phone = order?.phone_number || {};
  const callingCode = phone.country_calling_code || config.countryCallingCode;
  const e164Digits = String(phone.e164_number || `+${callingCode}${phone.national_number || ''}`).replace(/\D/g, '');
  return {
    e164_number: `+${e164Digits}`,
    country_calling_code: callingCode,
    national_number: phone.national_number || e164Digits.replace(new RegExp(`^${callingCode}`), ''),
    country_iso2: phone.country_iso2 || config.countryISO2,
  };
}

function summarizeOrder(order, showPhone) {
  const phone = order?.phone_number || {};
  return {
    order_id: order?.order_id,
    status: order?.status,
    phone: redactPhone(phone.e164_number || `+${phone.country_calling_code || ''}${phone.national_number || ''}`, showPhone),
    country: phone.country_iso2,
    price: order?.price?.amount_decimal,
    expires_at: order?.expires_at,
  };
}

function logEvent(type, payload) {
  console.log(JSON.stringify({type, at: new Date().toISOString(), ...payload}));
}

async function listOffers(config) {
  const query = new URLSearchParams({
    application_key: config.applicationKey,
    country_iso2: config.countryISO2,
    country_calling_code: config.countryCallingCode,
    provider_key: config.providerKey,
  });
  const {body} = await jsonFetch(`${config.smsBaseURL}/price-offers?${query}`, {}, 60000);
  if (body.error) throw new Error(body.error.message || 'list SMSBower offers failed');
  return (body.offers || [])
    .map((offer) => ({
      provider_key: offer.provider_key || config.providerKey,
      provider_id: offer.upstream_provider_id || offer.offer_ref?.upstream_provider_id || 'default',
      price: Number(offer.price?.amount_decimal || Number.POSITIVE_INFINITY),
      count: Number(offer.available_count || 0),
      offer_ref: offer.offer_ref,
    }))
    .filter((offer) => offer.offer_ref)
    .filter((offer) => !config.providerID || offer.provider_id === config.providerID)
    .filter((offer) => offer.count >= config.minAvailableCount)
    .filter((offer) => config.price === undefined || Math.abs(offer.price - config.price) < 0.000001)
    .filter((offer) => config.minPrice === undefined || offer.price >= config.minPrice)
    .filter((offer) => config.maxPrice === undefined || offer.price <= config.maxPrice)
    .sort((a, b) => a.price - b.price || b.count - a.count);
}

async function acquireSMSNumber(config, offer, runID) {
  const body = {
    request_id: `${runID}-sms-${offer.provider_id}`,
    lease_duration: `${config.otpWaitSeconds + 300}s`,
    acquire_params: {
      offer_ref: offer.offer_ref,
      application_key: config.applicationKey,
      country_iso2: config.countryISO2,
      country_calling_code: config.countryCallingCode,
    },
  };
  return jsonFetch(`${config.smsBaseURL}/orders/acquire?wait_seconds=${config.acquireWaitSeconds}`, {
    method: 'POST',
    body: JSON.stringify(body),
  }, (config.acquireWaitSeconds + 30) * 1000);
}

async function cancelSMSOrder(config, orderID, reason, runID, showPhone) {
  if (!orderID) return undefined;
  const {status, body} = await jsonFetch(`${config.smsBaseURL}/orders/${encodeURIComponent(orderID)}/cancel`, {
    method: 'POST',
    body: JSON.stringify({request_id: `${runID}-cancel-${Date.now()}`, reason}),
  }, 30000).catch((error) => ({status: 0, body: {error: String(error)}}));
  const order = body.order?.order || body.order;
  return {status, error: body.error || body.error_message, order: summarizeOrder(order, showPhone)};
}

async function generateSharedPhoneCheckProxy(config, phone, runID) {
  const {status, body} = await jsonFetch(`${trimRightSlash(config.gopayActionBaseURL)}/gopay-toolbox/generate-shared-phone-check-proxy`, {
    method: 'POST',
    body: JSON.stringify({
      n8n_execution_id: runID,
      operation: 'register_indonesia_wa',
      country_code: 'US',
      data: {
        proxy_country_code: 'US',
        phone_country_calling_code: phone.country_calling_code,
        country_calling_code: phone.country_calling_code,
        country_iso2: phone.country_iso2,
      },
    }),
  }, 45000);
  return {
    status,
    success: body.success === true,
    state_json: body.state_json || '',
    error_message: body.error_message || body.error?.message || '',
    proxy_hash: body.data?.proxy_hash || '',
  };
}

async function releaseSharedPhoneCheckProxy(config, runID, stateJSON) {
  if (!stateJSON) return;
  await jsonFetch(`${trimRightSlash(config.gopayActionBaseURL)}/gopay-toolbox/release-shared-phone-check-proxy`, {
    method: 'POST',
    body: JSON.stringify({n8n_execution_id: runID, operation: 'register_indonesia_wa', state_json: stateJSON}),
  }, 15000).catch(() => undefined);
}

async function precheckNumber(config, phone, proxyStateJSON) {
  const waPromise =
    jsonFetch(`${config.waBaseURL}/phone/sms-probe`, {
      method: 'POST',
      body: JSON.stringify({workspace_id: 'default', request_id: requestID('wa-probe'), phone, proxy_state_json: proxyStateJSON}),
    }, 80000);
  const gopayPromise = config.skipGopayCheck
    ? Promise.resolve({status: 200, body: {success: true, available: true, status: 'skipped'}})
    : jsonFetch(`${config.gopayBaseURL}/phone/check`, {
      method: 'POST',
      body: JSON.stringify({phone: phone.national_number, country_code: `+${phone.country_calling_code}`, state_json: proxyStateJSON}),
    }, 80000);
  const [wa, gopay] = await Promise.all([waPromise, gopayPromise]);
  const waStatus = wa.body.phone_status || {};
  const waOK = wa.body.success === true && wa.body.request_failed !== true && waStatus.blocked !== true && waStatus.registered !== true && waStatus.sms_available === true;
  const gopayOK = config.skipGopayCheck || (gopay.body.success === true && gopay.body.available === true && gopay.body.status === 'available');
  return {
    ok: waOK && gopayOK,
    waOK,
    gopayOK,
    wa: {
      success: wa.body.success,
      request_failed: wa.body.request_failed,
      flow: waStatus.account_flow,
      registered: waStatus.registered,
      blocked: waStatus.blocked,
      sms_available: waStatus.sms_available,
      sms_wait_seconds: waStatus.sms_wait_seconds,
      methods: waStatus.method_statuses || [],
    },
    gopay: {
      success: gopay.body.success,
      available: gopay.body.available,
      status: gopay.body.status,
      error_message: gopay.body.error_message,
      skipped: config.skipGopayCheck,
    },
  };
}

async function createWAAccount(config, phone, runID) {
  const fingerprint = await jsonFetch(`${config.waBaseURL}/actions/fingerprints/random`, {
    method: 'POST',
    body: JSON.stringify({workspace_id: 'default', request_id: `${runID}-fp`, phone}),
  }, 30000);
  if (!fingerprint.body.success) throw new Error(fingerprint.body.error_message || 'generate WA fingerprint failed');
  const commit = await jsonFetch(`${config.waBaseURL}/actions/fingerprints/commit`, {
    method: 'POST',
    body: JSON.stringify({
      workspace_id: 'default',
      request_id: `${runID}-commit`,
      phone,
      transient_fingerprint_ref: fingerprint.body.transient_fingerprint_ref,
    }),
  }, 30000);
  if (!commit.body.success) throw new Error(commit.body.error_message || 'commit WA fingerprint failed');
  return commit.body;
}

async function cleanupWAAccount(config, waAccountID, runID) {
  if (!waAccountID) return undefined;
  const {status, body} = await jsonFetch(`${config.waBaseURL}/actions/registration/cleanup-failed-account`, {
    method: 'POST',
    body: JSON.stringify({workspace_id: 'default', request_id: `${runID}-wa-cleanup-${Date.now()}`, wa_account_id: waAccountID}),
  }, 30000).catch((error) => ({status: 0, body: {error_message: String(error)}}));
  return {status, success: body.success, deleted: body.deleted, error_message: body.error_message};
}

async function requestWAOTP(config, account, proxyStateJSON, runID) {
  return jsonFetch(`${config.waBaseURL}/actions/registration/request-sms-otp`, {
    method: 'POST',
    body: JSON.stringify({
      workspace_id: 'default',
      request_id: `${runID}-wa-code`,
      wa_account_id: account.wa_account_id,
      client_profile_id: account.client_profile_id,
      protocol_profile_id: account.protocol_profile_id,
      proxy_state_json: proxyStateJSON,
    }),
  }, 80000);
}

async function pollSMSCode(config, orderID, waitSeconds) {
  const deadline = Date.now() + waitSeconds * 1000;
  let polls = 0;
  while (Date.now() < deadline) {
    await delay(5000);
    polls += 1;
    const {body} = await jsonFetch(`${config.smsBaseURL}/order-codes?order_id=${encodeURIComponent(orderID)}&limit_per_order=5`, {}, 30000);
    for (const item of body.codes || []) {
      if (item.order_id === orderID && item.code?.value) return {code: item.code.value, after_seconds: polls * 5};
    }
    if (polls % 6 === 0) logEvent('otp_poll', {order_id: orderID, polls, code_received: false});
  }
  return {code: '', after_seconds: polls * 5};
}

async function submitWAOTP(config, verificationRequestID, code, proxyStateJSON, runID) {
  return jsonFetch(`${config.waBaseURL}/actions/registration/submit-otp`, {
    method: 'POST',
    body: JSON.stringify({workspace_id: 'default', request_id: `${runID}-wa-submit`, verification_request_id: verificationRequestID, code, proxy_state_json: proxyStateJSON}),
  }, 80000);
}

async function runAttempt(config, offer, runID, attempt) {
  let orderID = '';
  let account;
  let sharedProxyStateJSON = '';
  try {
    const acquire = await acquireSMSNumber(config, offer, runID);
    const order = acquire.body.order;
    orderID = order?.order_id || '';
    logEvent('sms_acquired', {attempt, offer: {provider: offer.provider_key, provider_id: offer.provider_id, price: offer.price, count: offer.count}, error: acquire.body.error, order: summarizeOrder(order, config.showPhone)});
    if (acquire.body.error || !orderID || !order?.phone_number) {
      const insufficientBalance = acquire.body.error?.code === 'SMS_ERROR_CODE_INSUFFICIENT_BALANCE';
      const cancel = await cancelSMSOrder(config, orderID, `sms_acquire_failed:${acquire.body.error?.message || ''}`, runID, config.showPhone);
      logEvent('sms_cancel_requested', {attempt, order_id: orderID, reason: 'sms_acquire_failed', cancel});
      return {success: false, reason: insufficientBalance ? 'sms_provider_insufficient_balance' : 'sms_acquire_failed', stop: insufficientBalance};
    }

    const phone = phoneFromOrder(order, config);
    const sharedProxy = await generateSharedPhoneCheckProxy(config, phone, runID);
    sharedProxyStateJSON = sharedProxy.state_json;
    logEvent('shared_detection_proxy_acquired', {attempt, order_id: orderID, success: sharedProxy.success, status: sharedProxy.status, proxy_hash: sharedProxy.proxy_hash, error_message: sharedProxy.error_message});
    if (!sharedProxy.success || !sharedProxyStateJSON) {
      const cancel = await cancelSMSOrder(config, orderID, `shared_proxy_failed:${sharedProxy.error_message || sharedProxy.status}`, runID, config.showPhone);
      logEvent('sms_cancel_requested', {attempt, order_id: orderID, reason: 'shared_proxy_failed', cancel});
      return {success: false, reason: 'shared_proxy_failed'};
    }

    const precheck = await precheckNumber(config, phone, sharedProxyStateJSON);
    logEvent('precheck', {attempt, order_id: orderID, phone: redactPhone(phone.e164_number, config.showPhone), ok: precheck.ok, wa: precheck.wa, gopay: precheck.gopay});
    if (!precheck.ok) {
      const cancel = await cancelSMSOrder(config, orderID, `precheck_failed wa_ok=${precheck.waOK} gopay_ok=${precheck.gopayOK}`, runID, config.showPhone);
      logEvent('sms_cancel_requested', {attempt, order_id: orderID, reason: 'precheck_failed', cancel});
      return {success: false, reason: 'precheck_failed'};
    }

    account = await createWAAccount(config, phone, runID);
    logEvent('wa_account_created', {attempt, order_id: orderID, wa_account_id: account.wa_account_id, client_profile_id: account.client_profile_id});

    logEvent('registration_proxy_reused', {attempt, order_id: orderID});

    const otpRequest = await requestWAOTP(config, account, sharedProxyStateJSON, runID);
    const verificationRequestID = otpRequest.body.verification_request_id || '';
    logEvent('wa_otp_requested', {attempt, order_id: orderID, success: otpRequest.body.success, status: otpRequest.body.status, verification_request_id: verificationRequestID, error_message: otpRequest.body.error_message});
    if (!otpRequest.body.success || !verificationRequestID) {
      const cancel = await cancelSMSOrder(config, orderID, `wa_otp_request_failed:${otpRequest.body.error_message || ''}`, runID, config.showPhone);
      const cleanup = await cleanupWAAccount(config, account.wa_account_id, runID);
      logEvent('attempt_failed_cleanup', {attempt, order_id: orderID, reason: 'wa_otp_request_failed', cancel, cleanup});
      return {success: false, reason: 'wa_otp_request_failed'};
    }

    const smsCode = await pollSMSCode(config, orderID, config.otpWaitSeconds);
    if (!smsCode.code) {
      const cancel = await cancelSMSOrder(config, orderID, 'otp_not_received', runID, config.showPhone);
      const cleanup = await cleanupWAAccount(config, account.wa_account_id, runID);
      logEvent('attempt_failed_cleanup', {attempt, order_id: orderID, reason: 'otp_not_received', after_seconds: smsCode.after_seconds, cancel, cleanup});
      return {success: false, reason: 'otp_not_received'};
    }
    logEvent('sms_code_received', {attempt, order_id: orderID, after_seconds: smsCode.after_seconds});

    const submit = await submitWAOTP(config, verificationRequestID, smsCode.code, sharedProxyStateJSON, runID);
    logEvent('wa_otp_submitted', {
      attempt,
      order_id: orderID,
      success: submit.body.success,
      status: submit.body.status,
      error_message: submit.body.error_message,
      registration_id: submit.body.registration?.registration_id,
      login_state_id: submit.body.login_state?.login_state_id,
      registered_identity_id: submit.body.login_state?.registered_identity_id,
    });
    if (!submit.body.success) {
      logEvent('attempt_paused_for_manual_retry', {
        attempt,
        order_id: orderID,
        reason: 'wa_submit_failed',
        wa_account_id: account.wa_account_id,
        verification_request_id: verificationRequestID,
      });
      return {success: false, reason: 'wa_submit_failed', stop: true, manual_retry: true, order_id: orderID, wa_account_id: account.wa_account_id, verification_request_id: verificationRequestID};
    }
    return {success: true, order_id: orderID, wa_account_id: account.wa_account_id};
  } finally {
    await releaseSharedPhoneCheckProxy(config, runID, sharedProxyStateJSON);
  }
}

async function main() {
  const config = parseArgs(process.argv.slice(2));
  const runID = requestID(`${config.providerKey}-${config.countryISO2.toLowerCase()}-wa`);
  const offers = await listOffers(config);
  if (!offers.length) throw new Error(`${config.providerKey} ID WA offer not found for requested policy`);
  logEvent('start', {run_id: runID, max_attempts: config.maxAttempts, min_available: config.minAvailableCount, offers: offers.slice(0, 5).map((offer) => ({provider: offer.provider_key, provider_id: offer.provider_id, price: offer.price, count: offer.count}))});
  for (let attempt = 1; attempt <= config.maxAttempts; attempt += 1) {
    const offer = offers[(attempt - 1) % offers.length];
    const result = await runAttempt(config, offer, `${runID}-${attempt}`, attempt);
    if (result.success) {
      logEvent('done', {run_id: runID, result});
      return;
    }
    logEvent('attempt_done', {run_id: runID, attempt, result});
    if (result.stop) {
      logEvent('done', {run_id: runID, result});
      process.exitCode = 1;
      return;
    }
    if (attempt < config.maxAttempts && config.attemptDelaySeconds > 0) {
      await delay(config.attemptDelaySeconds * 1000);
    }
  }
  logEvent('done', {run_id: runID, result: {success: false, reason: 'max_attempts_exhausted'}});
  process.exitCode = 1;
}

main().catch((error) => {
  logEvent('fatal', {error: String(error?.stack || error)});
  process.exitCode = 1;
});
