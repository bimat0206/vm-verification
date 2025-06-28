#!/usr/bin/env node

// Script to check verification status and recent verifications
const API_BASE_URL = 'https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1';
const API_KEY = 'WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH';

console.log('🔍 Checking verification status...\n');

async function checkRecentVerifications() {
  console.log('1. Fetching recent verifications...');
  try {
    const response = await fetch(`${API_BASE_URL}/api/verifications?limit=10&sortBy=newest`, {
      method: 'GET',
      headers: {
        'X-Api-Key': API_KEY,
        'Accept': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    console.log(`   Found ${data.results?.length || 0} recent verifications:`);
    
    if (data.results && data.results.length > 0) {
      data.results.forEach((verification, index) => {
        const createdAt = new Date(verification.verificationAt || verification.createdAt);
        const now = new Date();
        const ageMinutes = Math.round((now.getTime() - createdAt.getTime()) / 60000);
        
        console.log(`   ${index + 1}. ID: ${verification.verificationId}`);
        console.log(`      Status: ${verification.verificationStatus}`);
        console.log(`      Type: ${verification.verificationType || 'N/A'}`);
        console.log(`      Age: ${ageMinutes} minutes ago`);
        console.log(`      Has Analysis: ${verification.llmAnalysis ? 'YES' : 'NO'}`);
        console.log(`      Machine ID: ${verification.vendingMachineId || 'N/A'}`);
        console.log('');
      });
    }
    
    // Check for stuck verifications (older than 15 minutes and still pending)
    const stuckVerifications = data.results?.filter(v => {
      const age = new Date().getTime() - new Date(v.verificationAt || v.createdAt).getTime();
      const ageMinutes = age / 60000;
      return ageMinutes > 15 && (v.verificationStatus === 'PENDING' || v.verificationStatus === 'PROCESSING');
    }) || [];
    
    if (stuckVerifications.length > 0) {
      console.log('⚠️  Found potentially stuck verifications:');
      stuckVerifications.forEach(v => {
        const age = Math.round((new Date().getTime() - new Date(v.verificationAt || v.createdAt).getTime()) / 60000);
        console.log(`   - ${v.verificationId} (${v.verificationStatus}, ${age} minutes old)`);
      });
    } else {
      console.log('✅ No stuck verifications found');
    }
    
  } catch (error) {
    console.log(`   ❌ Failed: ${error.message}`);
  }
}

async function checkBackendHealth() {
  console.log('\n2. Checking backend health...');
  try {
    const response = await fetch(`${API_BASE_URL}/api/health`, {
      method: 'GET',
      headers: {
        'X-Api-Key': API_KEY,
        'Accept': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    const health = await response.json();
    console.log(`   Status: ${health.status}`);
    console.log(`   Services:`);
    
    if (health.services) {
      Object.entries(health.services).forEach(([service, info]) => {
        console.log(`     ${service}: ${info.status} - ${info.message}`);
      });
    }
    
    console.log('   ✅ Backend is healthy');
    
  } catch (error) {
    console.log(`   ❌ Backend health check failed: ${error.message}`);
  }
}

async function checkSpecificVerification(verificationId) {
  console.log(`\n3. Checking specific verification: ${verificationId}...`);
  try {
    const response = await fetch(`${API_BASE_URL}/api/verifications?verificationId=${verificationId}&limit=1`, {
      method: 'GET',
      headers: {
        'X-Api-Key': API_KEY,
        'Accept': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    
    if (data.results && data.results.length > 0) {
      const verification = data.results[0];
      console.log(`   Found verification:`);
      console.log(`     ID: ${verification.verificationId}`);
      console.log(`     Status: ${verification.verificationStatus}`);
      console.log(`     Type: ${verification.verificationType}`);
      console.log(`     Created: ${verification.verificationAt || verification.createdAt}`);
      console.log(`     Has Analysis: ${verification.llmAnalysis ? 'YES' : 'NO'}`);
      console.log(`     Overall Accuracy: ${verification.overallAccuracy || 'N/A'}`);
      
      if (verification.rawData) {
        console.log(`     Raw Data Keys: ${Object.keys(verification.rawData).join(', ')}`);
      }
    } else {
      console.log(`   ❌ Verification ${verificationId} not found`);
    }
    
  } catch (error) {
    console.log(`   ❌ Failed to check verification: ${error.message}`);
  }
}

// Main execution
async function main() {
  await checkRecentVerifications();
  await checkBackendHealth();
  
  // If a verification ID is provided as argument, check it specifically
  const verificationId = process.argv[2];
  if (verificationId) {
    await checkSpecificVerification(verificationId);
  } else {
    console.log('\n💡 To check a specific verification, run:');
    console.log('   node check-verification-status.js <verification-id>');
  }
  
  console.log('\n🏁 Status check completed!');
}

main().catch(console.error);
