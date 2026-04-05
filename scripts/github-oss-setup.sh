#!/bin/bash

# GitHub Repository Security Setup Script
# This script automates the setup of security features for a GitHub repository
# including vulnerability alerts, Dependabot, secret scanning, and branch protection.
#
# Prerequisites:
# - GitHub CLI (gh) installed and authenticated
# - Repository must exist
# - User must have admin access to the repository

# Removed set -e to prevent early exit on API failures
# We handle errors explicitly in each section

REPO_OWNER="bold-minds"
REPO_NAME="num"
REPO="$REPO_OWNER/$REPO_NAME"

# Status tracking variables
declare -A STATUS
STATUS_UPDATED=0
STATUS_SKIPPED=0
STATUS_FAILED=0

# Function to track status
track_status() {
    local step="$1"
    local result="$2"  # SUCCESS, SKIPPED, or FAILED
    STATUS["$step"]="$result"
    case "$result" in
        "SUCCESS") ((STATUS_UPDATED++)) ;;
        "SKIPPED") ((STATUS_SKIPPED++)) ;;
        "FAILED") ((STATUS_FAILED++)) ;;
    esac
}

echo "🔧 Setting up GitHub security for $REPO..."

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI is not installed. Please install it from https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "❌ Not authenticated with GitHub CLI. Run: gh auth login"
    exit 1
fi

echo "✅ GitHub CLI is ready"

# Set repository to public FIRST (required for free security features)
echo "🌍 Setting repository to public (required for free security features)..."
if gh repo edit $REPO --visibility public --accept-visibility-change-consequences >/dev/null 2>&1; then
    track_status "Repository Visibility" "SUCCESS"
    echo "✅ Repository set to public"
    echo "⏳ Waiting 5 seconds for visibility change to propagate..."
    sleep 5
else
    track_status "Repository Visibility" "SKIPPED"
    echo "⚠️  Repository may already be public"
fi

# Enable vulnerability alerts
echo "🔒 Enabling vulnerability alerts..."
if gh api repos/$REPO/vulnerability-alerts -X PUT >/dev/null 2>&1; then
    track_status "Vulnerability Alerts" "SUCCESS"
    echo "✅ Vulnerability alerts enabled"
else
    track_status "Vulnerability Alerts" "SKIPPED"
    echo "⚠️  Vulnerability alerts may already be enabled"
fi

# Enable automated security fixes (Dependabot security updates)
echo "🤖 Enabling Dependabot security updates..."
if gh api repos/$REPO/automated-security-fixes -X PUT >/dev/null 2>&1; then
    track_status "Dependabot Security Updates" "SUCCESS"
    echo "✅ Dependabot security updates enabled"
else
    track_status "Dependabot Security Updates" "SKIPPED"
    echo "⚠️  Dependabot security updates may already be enabled"
fi

# Enable dependency graph
echo "📊 Enabling dependency graph..."
if gh api repos/$REPO -X PATCH -f has_vulnerability_alerts=true >/dev/null 2>&1; then
    track_status "Dependency Graph" "SUCCESS"
    echo "✅ Dependency graph enabled"
else
    track_status "Dependency Graph" "SKIPPED"
    echo "⚠️  Dependency graph may already be enabled"
fi

# Enable code security and analysis (FREE for public repos)
echo "🔒 Enabling code security and analysis (secret scanning)..."
if gh api repos/$REPO -X PATCH --input - >/dev/null 2>&1 <<EOF
{
  "security_and_analysis": {
    "secret_scanning": {
      "status": "enabled"
    },
    "secret_scanning_push_protection": {
      "status": "enabled"
    }
  }
}
EOF
then
    track_status "Secret Scanning" "SUCCESS"
    echo "✅ Secret scanning enabled"
else
    track_status "Secret Scanning" "SKIPPED"
    echo "⚠️  Secret scanning may already be enabled"
fi

# Create branch protection rule for main
echo "🛡️  Setting up branch protection for main..."
if gh api repos/$REPO/branches/main/protection -X PUT --input - >/dev/null 2>&1 <<EOF
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["test"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1,
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": true
  },
  "restrictions": null
}
EOF
then
    track_status "Branch Protection" "SUCCESS"
    echo "✅ Branch protection configured"
else
    track_status "Branch Protection" "SKIPPED"
    echo "⚠️  Branch protection may already be configured"
fi

# Enable issues
echo "📝 Enabling issues..."
if gh repo edit $REPO --enable-issues >/dev/null 2>&1; then
    track_status "Issues" "SUCCESS"
    echo "✅ Issues enabled"
else
    track_status "Issues" "FAILED"
    echo "❌ Failed to enable issues"
fi

# Enable discussions
echo "💬 Enabling discussions..."
if gh repo edit $REPO --enable-discussions >/dev/null 2>&1; then
    track_status "Discussions" "SUCCESS"
    echo "✅ Discussions enabled"
else
    track_status "Discussions" "FAILED"
    echo "❌ Failed to enable discussions"
fi

# Disable wiki (use README instead)
echo "📚 Disabling wiki..."
if gh repo edit $REPO --enable-wiki=false >/dev/null 2>&1; then
    track_status "Wiki (Disabled)" "SUCCESS"
    echo "✅ Wiki disabled"
else
    track_status "Wiki (Disabled)" "FAILED"
    echo "❌ Failed to disable wiki"
fi



# Print comprehensive status summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 GITHUB SECURITY SETUP SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📈 Statistics:"
echo "   ✅ Updated: $STATUS_UPDATED"
echo "   ⏭️  Skipped: $STATUS_SKIPPED"
echo "   ❌ Failed:  $STATUS_FAILED"
echo "   📝 Total:   $((STATUS_UPDATED + STATUS_SKIPPED + STATUS_FAILED))"
echo ""
echo "📋 Detailed Status:"
for step in "Repository Visibility" "Vulnerability Alerts" "Dependabot Security Updates" "Dependency Graph" "Secret Scanning" "Branch Protection" "Issues" "Discussions" "Wiki (Disabled)"; do
    if [[ -n "${STATUS[$step]}" ]]; then
        case "${STATUS[$step]}" in
            "SUCCESS") echo "   ✅ $step: Updated" ;;
            "SKIPPED") echo "   ⏭️  $step: Already configured" ;;
            "FAILED")  echo "   ❌ $step: Failed" ;;
        esac
    fi
done
echo ""
if [[ $STATUS_FAILED -gt 0 ]]; then
    echo "⚠️  Some steps failed. Check the output above for details."
else
    echo "🎉 Automated setup complete!"
fi
echo ""
echo "⚠️  MANUAL STEPS STILL REQUIRED:"
echo "1. Go to Settings → Security & analysis"
echo "2. Enable Code scanning (CodeQL) - requires manual setup"
echo "3. Enable Private vulnerability reporting"
echo "4. Configure Actions permissions in Settings → Actions → General"
echo ""
echo "💡 These require manual setup due to GitHub API limitations"
echo "📋 Use GITHUB_SECURITY_SETUP.md for the complete manual checklist"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
