/// A simple hello world library.
library hello_world;

/// Returns a hello world message.
/// 
/// Example:
/// ```dart
/// print(helloWorld()); // Hello, World!
/// ```
String helloWorld() {
  return 'Hello, World!';
}

/// Returns a personalized hello message.
/// 
/// [name] The name to greet.
/// 
/// Example:
/// ```dart
/// print(helloPersonalized('Alice')); // Hello, Alice!
/// ```
String helloPersonalized(String name) {
  return 'Hello, $name!';
}