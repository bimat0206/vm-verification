// Simple direct test that bypasses SAM
process.env.USE_MOCK_S3 = 'true';
process.env.AWS_SAM_LOCAL = 'true';
process.env.TEMP_DIR = '/tmp/s3-mock';

const { handler } = require('./index');
const event = require('./test-event.json');

console.log('Starting direct test...');
handler(event)
  .then(result => {
    console.log('Test completed successfully:', result);
    process.exit(0);
  })
  .catch(error => {
    console.error('Test failed:', error);
    process.exit(1);
  });
