#!/bin/bash

# ==========================================
# Fuli Robust Integration Test: Nested Negation
# ==========================================

# 1. SETUP & BUILD
# ------------------------------------------
# Define paths
TEST_DIR="test_recursive_negation"
BINARY="./fuli_test_bin"

echo "🔨 Building Fuli from source..."
go build -o "$BINARY" main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed. Aborting."
    exit 1
fi

# Clean previous test artifacts
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR/logs"

# 2. CREATE MOCK CONTENT
# ------------------------------------------
echo "I am trash" > "$TEST_DIR/logs/trash.log"
echo "I am treasure" > "$TEST_DIR/logs/treasure.log"
echo "I am root" > "$TEST_DIR/root.txt"

# 3. DEFINE TESTS
# ------------------------------------------
FAIL=0

run_test() {
    local name="$1"
    local ignore_content="$2"
    local expected_file="$3"
    local unexpected_file="$4"
    local output_file="$TEST_DIR/${name}_output.txt"

    echo "---------------------------------------------------"
    echo "Testing Scenario: $name"
    
    # Write the ignore file
    echo "$ignore_content" > "$TEST_DIR/.contextignore"

    # Run Fuli
    # Note: We do NOT suppress stderr (2>&1) so we can see panic/errors if they happen.
    "$BINARY" -o "$output_file" "$TEST_DIR"
    
    # A. Check for Application Crash/Failure
    if [ $? -ne 0 ]; then
        echo "❌ CRITICAL: Application exited with error code."
        FAIL=1
        return
    fi

    # B. Check for Output Generation
    if [ ! -f "$output_file" ]; then
        echo "❌ CRITICAL: Output file was not created."
        FAIL=1
        return
    fi

    # C. Validate Content (Positive Assertion)
    if grep -q "$expected_file" "$output_file"; then
        echo "✅ PASS: Found expected file '$expected_file'"
    else
        echo "❌ FAIL: Missing expected file '$expected_file'"
        FAIL=1
    fi

    # D. Validate Content (Negative Assertion)
    if grep -q "$unexpected_file" "$output_file"; then
        echo "❌ FAIL: Found unexpected file '$unexpected_file' (Should be ignored)"
        FAIL=1
    else
        echo "✅ PASS: Correctly ignored '$unexpected_file'"
    fi
}

# 4. EXECUTE SCENARIOS
# ------------------------------------------

# Scenario A: The "Naive" Approach (Git/Fuli Standard Behavior)
# If we ignore the parent dir 'logs/', we cannot re-include files inside it.
run_test "naive_approach" \
$'logs/\n!logs/treasure.log' \
"root.txt" \
"treasure.log" 
# Expectation: treasure.log is MISSING because logs/ is skipped.

# Scenario B: The "Correct" Approach (Wildcard Content)
# We ignore logs/* (contents), allowing us to traverse into logs/ and negate specific files.
run_test "correct_approach" \
$'logs/*\n!logs/treasure.log' \
"treasure.log" \
"trash.log"
# Expectation: treasure.log is FOUND. trash.log is IGNORED.

# 5. CLEANUP
# ------------------------------------------
echo "---------------------------------------------------"
rm "$BINARY"
if [ $FAIL -eq 0 ]; then
    echo "🎉 ALL TESTS PASSED"
    rm -rf "$TEST_DIR"
    exit 0
else
    echo "💥 SOME TESTS FAILED"
    # Keep directory for inspection
    exit 1
fi