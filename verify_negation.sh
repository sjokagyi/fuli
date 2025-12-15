#!/bin/bash

# 1. Setup Playground
rm -rf test_negation
mkdir -p test_negation
cd test_negation

# 2. Create Files
echo "This is secret" > db.secret
echo "This is public" > public.secret
echo "This should be ignored" > other.secret

# 3. Create .contextignore
# We ignore all .secret files, but explicitly ALLOW public.secret
cat <<EOF > .contextignore
*.secret
!public.secret
EOF

# 4. Run Fuli (assuming fuli is installed/built in parent dir)
# We expect only public.secret (and .contextignore) to be copied.
../fuli -o output.txt .

# 5. Verify Output
echo "--- Verification ---"
if grep -q "db.secret" output.txt; then
    echo "❌ FAIL: db.secret was found (should be ignored)"
else
    echo "✅ PASS: db.secret was ignored"
fi

if grep -q "public.secret" output.txt; then
    echo "✅ PASS: public.secret was found (Negation worked)"
else
    echo "❌ FAIL: public.secret was missing (Negation failed)"
fi

# Clean up
cd ..
# rm -rf test_negation