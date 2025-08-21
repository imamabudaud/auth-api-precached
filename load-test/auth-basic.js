import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },
    { duration: '1m', target: 100 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.1'],
  },
};

export default function () {
  // Generate random username from 1-1000000 with 8-digit zero padding
  const randomId = Math.floor(Math.random() * 1000000) + 1;
  const username = randomId.toString().padStart(8, '0') + '@katakode.com';
  
  const payload = JSON.stringify({
    username: username,
    password: 'testpassword123'
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post('http://localhost:8080/login', payload, params);

  // Log some requests for debugging (only occasionally to avoid spam)
  if (Math.random() < 0.01) { // 10% chance to log
    console.log(`Testing username: ${username}, Status: ${response.status}`);
  }

  check(response, {
    'status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
