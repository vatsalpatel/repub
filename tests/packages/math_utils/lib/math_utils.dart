/// A collection of mathematical utility functions.
library math_utils;

/// Checks if a number is prime.
/// 
/// Returns `true` if [n] is a prime number, `false` otherwise.
bool isPrime(int n) {
  if (n <= 1) return false;
  if (n <= 3) return true;
  if (n % 2 == 0 || n % 3 == 0) return false;
  
  for (int i = 5; i * i <= n; i += 6) {
    if (n % i == 0 || n % (i + 2) == 0) return false;
  }
  
  return true;
}

/// Calculates the factorial of a number.
/// 
/// Returns the factorial of [n]. Throws [ArgumentError] if [n] is negative.
int factorial(int n) {
  if (n < 0) throw ArgumentError('Factorial is not defined for negative numbers');
  if (n <= 1) return 1;
  
  int result = 1;
  for (int i = 2; i <= n; i++) {
    result *= i;
  }
  
  return result;
}

/// Calculates the greatest common divisor of two numbers.
/// 
/// Returns the GCD of [a] and [b].
int gcd(int a, int b) {
  a = a.abs();
  b = b.abs();
  
  while (b != 0) {
    int temp = b;
    b = a % b;
    a = temp;
  }
  
  return a;
}

/// Calculates the least common multiple of two numbers.
/// 
/// Returns the LCM of [a] and [b].
int lcm(int a, int b) {
  if (a == 0 || b == 0) return 0;
  return (a.abs() * b.abs()) ~/ gcd(a, b);
}