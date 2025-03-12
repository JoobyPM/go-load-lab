import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  // Ramp up load (via number of VUs) in stages
  scenarios: {
    gradual_load_test: {
      executor: 'ramping-vus',
      startVUs: 5,
      stages: [
        // Stage 1: ramp up to 50 VUs over 1 minutes
        { target: 50, duration: '5m' },
        // Stage 2: ramp up to 100 VUs over 1 minutes
        { target: 100, duration: '5m' },
        // Stage 3: keep 100 VUs for 1 minutes
        { target: 100, duration: '5m' },
        // Stage 4: keep 200 VUs for 1 minutes
        { target: 200, duration: '15m' },
        
        
      ],
      gracefulRampDown: '30s',
    }
  },
  thresholds: {
    http_req_duration: ['max<=10000'], // If max > 1000ms, fail the test
  },
};

export default function () {
  // It's an example of a ramp-up test, where the number of VUs is increased gradually.
  // Replace with your MetalLB IP or Ingress address
  const url = 'http://127.0.0.1:8080/items?offset=90000&limit=1000';
  const res = http.get(url);

  // Optionally check for success & measure response times
  check(res, {
    'status is 200': (r) => r.status === 200,
    'latency <= 1000ms': (r) => r.timings.duration <= 1000,
  });

  sleep(0.01); // small wait before next iteration, to avoid overwhelming the server
}
