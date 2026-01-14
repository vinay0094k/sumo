import { CognitoUserPool, CognitoUser, AuthenticationDetails, CognitoUserAttribute } from 'amazon-cognito-identity-js';
import { AWS_REGION, USER_POOL_ID, USER_POOL_WEB_CLIENT_ID } from '@/env';

const poolData = {
  UserPoolId: USER_POOL_ID,
  ClientId: USER_POOL_WEB_CLIENT_ID,
};

const userPool = new CognitoUserPool(poolData);

export const cognitoConfig = {
  region: AWS_REGION,
  userPoolId: USER_POOL_ID,
  userPoolWebClientId: USER_POOL_WEB_CLIENT_ID,
};

export default userPool;
