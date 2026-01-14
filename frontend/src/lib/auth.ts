import { CognitoUser, AuthenticationDetails, CognitoUserAttribute } from 'amazon-cognito-identity-js';
import userPool from './cognito';
import type { User } from './types';

const TOKEN_KEY = 'auth_token';

export const getToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  const cognitoUser = userPool.getCurrentUser();
  if (!cognitoUser) return null;
  
  let token: string | null = null;
  cognitoUser.getSession((err: any, session: any) => {
    if (!err && session) {
      token = session.getIdToken().getJwtToken();
    }
  });
  return token;
};

export const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, token);
};

export const removeToken = (): void => {
  localStorage.removeItem(TOKEN_KEY);
  const cognitoUser = userPool.getCurrentUser();
  if (cognitoUser) {
    cognitoUser.signOut();
  }
};

export const login = async (email: string, password: string): Promise<{ token: string; user: User }> => {
  return new Promise((resolve, reject) => {
    const authenticationDetails = new AuthenticationDetails({
      Username: email,
      Password: password,
    });

    const cognitoUser = new CognitoUser({
      Username: email,
      Pool: userPool,
    });

    cognitoUser.authenticateUser(authenticationDetails, {
      onSuccess: (session) => {
        const token = session.getIdToken().getJwtToken();
        const payload = session.getIdToken().payload;
        
        const user: User = {
          id: payload.sub,
          email: payload.email,
          name: payload.name || email.split('@')[0],
        };

        setToken(token);
        resolve({ token, user });
      },
      onFailure: (err) => {
        reject(new Error(err.message || 'Login failed'));
      },
    });
  });
};

export const signup = async (email: string, password: string, name: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    const attributeList = [
      new CognitoUserAttribute({ Name: 'email', Value: email }),
      new CognitoUserAttribute({ Name: 'name', Value: name }),
    ];

    userPool.signUp(email, password, attributeList, [], (err, result) => {
      if (err) {
        reject(new Error(err.message || 'Signup failed'));
        return;
      }
      resolve();
    });
  });
};

export const confirmSignup = async (email: string, code: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    const cognitoUser = new CognitoUser({
      Username: email,
      Pool: userPool,
    });

    cognitoUser.confirmRegistration(code, true, (err, result) => {
      if (err) {
        reject(new Error(err.message || 'Confirmation failed'));
        return;
      }
      resolve();
    });
  });
};

export const getCurrentUser = (): Promise<User | null> => {
  return new Promise((resolve) => {
    const cognitoUser = userPool.getCurrentUser();
    if (!cognitoUser) {
      resolve(null);
      return;
    }

    cognitoUser.getSession((err: any, session: any) => {
      if (err || !session) {
        resolve(null);
        return;
      }

      const payload = session.getIdToken().payload;
      resolve({
        id: payload.sub,
        email: payload.email,
        name: payload.name || payload.email.split('@')[0],
      });
    });
  });
};

export const isAuthenticated = (): boolean => {
  return !!getToken();
};
